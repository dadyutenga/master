package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/views/auth"
	"github.com/dadyutenga/hms-control/internal/views/home"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Home(c *fiber.Ctx) error {
	return render(c, home.Welcome())
}

func (h *Handler) ShowRegister(c *fiber.Ctx) error {
	return render(c, auth.RegisterStep1(auth.RegisterProps{}))
}

func (h *Handler) ShowLogin(c *fiber.Ctx) error {
	verified := c.Query("verified") == "1"
	return render(c, auth.Login(auth.LoginProps{Verified: verified, Error: ""}))
}

// ── Multi-Step Registration ─────────────────────────────────────────────────

func (h *Handler) RegisterStep1(c *fiber.Ctx) error {
	tin := strings.TrimSpace(c.FormValue("tin"))
	brela := strings.TrimSpace(c.FormValue("brela_number"))
	hotelName := strings.TrimSpace(c.FormValue("hotel_name"))
	category := strings.TrimSpace(c.FormValue("category"))
	roomCountStr := strings.TrimSpace(c.FormValue("room_count"))
	address := strings.TrimSpace(c.FormValue("address"))
	city := strings.TrimSpace(c.FormValue("city"))
	country := strings.TrimSpace(c.FormValue("country"))

	if tin == "" || brela == "" || hotelName == "" || category == "" || roomCountStr == "" || address == "" || city == "" || country == "" {
		return render(c, auth.RegisterStep1(auth.RegisterProps{Error: "All fields are required."}))
	}
	roomCount, err := strconv.ParseInt(roomCountStr, 10, 64)
	if err != nil || roomCount < 1 {
		return render(c, auth.RegisterStep1(auth.RegisterProps{Error: "Room count must be a positive number."}))
	}

	sess, _ := h.store.Get(c)
	sess.Set("reg_tin", tin)
	sess.Set("reg_brela", brela)
	sess.Set("reg_hotel_name", hotelName)
	sess.Set("reg_category", category)
	sess.Set("reg_room_count", roomCount)
	sess.Set("reg_address", address)
	sess.Set("reg_city", city)
	sess.Set("reg_country", country)
	if err := sess.Save(); err != nil {
		return render(c, auth.RegisterStep1(auth.RegisterProps{Error: "Session error."}))
	}
	return render(c, auth.RegisterStep2(auth.RegisterProps{}))
}

func (h *Handler) RegisterStep2(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.FormValue("name"))
	phone := strings.TrimSpace(c.FormValue("phone"))
	email := strings.TrimSpace(c.FormValue("email"))
	pass := c.FormValue("password")
	confirm := c.FormValue("password_confirmation")

	if name == "" || email == "" || pass == "" {
		return render(c, auth.RegisterStep2(auth.RegisterProps{Error: "Name, email, and password are required."}))
	}
	if pass != confirm {
		return render(c, auth.RegisterStep2(auth.RegisterProps{Error: "Passwords do not match."}))
	}
	if len(pass) < 8 {
		return render(c, auth.RegisterStep2(auth.RegisterProps{Error: "Password must be at least 8 characters."}))
	}

	q := generated.New(h.db)
	_, err := q.GetUserByEmail(c.Context(), email)
	if err == nil {
		return render(c, auth.RegisterStep2(auth.RegisterProps{Error: "Email already registered."}))
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

	sess, _ := h.store.Get(c)
	sess.Set("reg_name", name)
	sess.Set("reg_email", email)
	if phone != "" {
		sess.Set("reg_phone", phone)
	}
	sess.Set("reg_password", string(hash))
	if err := sess.Save(); err != nil {
		return render(c, auth.RegisterStep2(auth.RegisterProps{Error: "Session error."}))
	}
	return render(c, auth.RegisterStep3(auth.RegisterProps{}))
}

func (h *Handler) RegisterStep3(c *fiber.Ctx) error {
	sess, _ := h.store.Get(c)
	email, _ := sess.Get("reg_email").(string)
	if email == "" {
		return c.Redirect("/register")
	}

	brelaFile, err := c.FormFile("brela_certificate")
	if err != nil {
		return render(c, auth.RegisterStep3(auth.RegisterProps{Error: "BRELA certificate is required."}))
	}
	traFile, err := c.FormFile("tra_certificate")
	if err != nil {
		return render(c, auth.RegisterStep3(auth.RegisterProps{Error: "TRA certificate is required."}))
	}

	if brelaFile.Size > h.cfg.MaxUploadSize || traFile.Size > h.cfg.MaxUploadSize {
		return render(c, auth.RegisterStep3(auth.RegisterProps{Error: "Files must be under 10MB each."}))
	}
	if !strings.HasPrefix(brelaFile.Header.Get("Content-Type"), "image/") ||
		!strings.HasPrefix(traFile.Header.Get("Content-Type"), "image/") {
		return render(c, auth.RegisterStep3(auth.RegisterProps{Error: "Only image files (JPEG, PNG) are accepted."}))
	}

	company, _ := sess.Get("reg_hotel_name").(string)
	name, _ := sess.Get("reg_name").(string)
	phone, _ := sess.Get("reg_phone").(string)
	password, _ := sess.Get("reg_password").(string)
	tin, _ := sess.Get("reg_tin").(string)
	brela, _ := sess.Get("reg_brela").(string)
	hotelName, _ := sess.Get("reg_hotel_name").(string)
	category, _ := sess.Get("reg_category").(string)
	roomCount, _ := sess.Get("reg_room_count").(int64)
	address, _ := sess.Get("reg_address").(string)
	city, _ := sess.Get("reg_city").(string)
	country, _ := sess.Get("reg_country").(string)

	q := generated.New(h.db)

	var phonePtr *string
	if phone != "" { phonePtr = &phone }
	var tinPtr, brelaPtr *string
	if tin != "" { tinPtr = &tin }
	if brela != "" { brelaPtr = &brela }

	user, err := q.CreateUser(c.Context(), generated.CreateUserParams{
		Name: name, Email: email, Company: company, Phone: phonePtr, Password: password, TIN: tinPtr, BrelaNumber: brelaPtr,
	})
	if err != nil {
		return render(c, auth.RegisterStep3(auth.RegisterProps{Error: "Registration failed."}))
	}

	slug := generateSlug(q, c.Context(), company)
	domain := slug + "." + h.cfg.BaseDomain
	dbPass := randomHex(16)
	var roomCountPtr *int64
	if roomCount > 0 { roomCountPtr = &roomCount }

	_, err = q.CreateTenant(c.Context(), generated.CreateTenantParams{
		UserID: user.ID, CompanyName: company, Slug: slug, Domain: domain,
		DbName: "hms_" + slug + "_db", DbUser: "hms_" + slug + "_user", DbPassword: dbPass,
		HotelName: &hotelName, Category: &category, RoomCount: roomCountPtr,
		Address: &address, City: &city, Country: &country,
	})
	if err != nil {
		return render(c, auth.RegisterStep3(auth.RegisterProps{Error: "Failed to create tenant."}))
	}

	uploadDir := filepath.Join(h.cfg.UploadDir, strconv.FormatInt(user.ID, 10))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return render(c, auth.RegisterStep3(auth.RegisterProps{Error: "Storage error."}))
	}

	brelaPath, err := saveFile(brelaFile, uploadDir, "brela_")
	if err != nil {
		return render(c, auth.RegisterStep3(auth.RegisterProps{Error: "Failed to save BRELA certificate."}))
	}
	q.CreateDocument(c.Context(), generated.CreateDocumentParams{
		UserID: user.ID, DocType: "brela_certificate", Filename: brelaPath,
		OriginalName: brelaFile.Filename, MimeType: brelaFile.Header.Get("Content-Type"), SizeBytes: brelaFile.Size,
	})

	traPath, err := saveFile(traFile, uploadDir, "tra_")
	if err != nil {
		return render(c, auth.RegisterStep3(auth.RegisterProps{Error: "Failed to save TRA certificate."}))
	}
	q.CreateDocument(c.Context(), generated.CreateDocumentParams{
		UserID: user.ID, DocType: "tra_certificate", Filename: traPath,
		OriginalName: traFile.Filename, MimeType: traFile.Header.Get("Content-Type"), SizeBytes: traFile.Size,
	})

	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)
	q.CreateVerifyToken(c.Context(), generated.CreateVerifyTokenParams{UserID: user.ID, Token: token})
	go h.mail.SendVerification(user.Email, user.Name, h.cfg.AppURL+"/verify/"+token)

	sess.Destroy()
	return c.Redirect("/verify-notice")
}

func (h *Handler) Register(c *fiber.Ctx) error {
	return c.Redirect("/register")
}

// ── Verify, Login, Logout ───────────────────────────────────────────────────

func (h *Handler) VerifyEmail(c *fiber.Ctx) error {
	token := c.Params("token")
	q := generated.New(h.db)

	row, err := q.GetVerifyToken(c.Context(), token)
	if err != nil {
		return c.Status(400).SendString("Invalid or expired verification link.")
	}
	q.VerifyUser(c.Context(), row.Uid)
	q.UseVerifyToken(c.Context(), token)

	tenant, _ := q.GetTenantByUserID(c.Context(), row.Uid)
	q.UpdateTenantStatus(c.Context(), generated.UpdateTenantStatusParams{
		ID: tenant.ID, Status: generated.TenantStatusPendingApproval,
	})
	return c.Redirect("/login?verified=1")
}

func (h *Handler) Login(c *fiber.Ctx) error {
	email := strings.TrimSpace(c.FormValue("email"))
	pass := c.FormValue("password")

	q := generated.New(h.db)
	user, err := q.GetUserByEmail(c.Context(), email)
	if err != nil {
		return render(c, auth.Login(auth.LoginProps{Error: "Invalid email or password."}))
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass)); err != nil {
		return render(c, auth.Login(auth.LoginProps{Error: "Invalid email or password."}))
	}
	if !user.Verified {
		return render(c, auth.Login(auth.LoginProps{Error: "Please verify your email first."}))
	}

	sess, _ := h.store.Get(c)
	sess.Set("userID", user.ID)
	sess.Set("role", user.Role)
	sess.Save()

	if user.Role == "superadmin" {
		return c.Redirect("/admin")
	}
	return c.Redirect("/dashboard")
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	sess, _ := h.store.Get(c)
	sess.Destroy()
	return c.Redirect("/login")
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func saveFile(fh *multipart.FileHeader, dir, prefix string) (string, error) {
	ext := filepath.Ext(fh.Filename)
	target := filepath.Join(dir, prefix+randomHex(8)+ext)

	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(target)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}
	return target, nil
}

func generateSlug(q *generated.Queries, ctx context.Context, company string) string {
	re := strings.NewReplacer(" ", "", "-", "", "_", "")
	base := strings.ToLower(re.Replace(company))
	if len(base) > 12 {
		base = base[:12]
	}
	slug := base
	i := 2
	for {
		_, err := q.GetTenantBySlug(ctx, slug)
		if err != nil {
			break
		}
		suffix := base
		if len(base) > 10 {
			suffix = base[:10]
		}
		slug = suffix + string(rune('0'+i))
		i++
	}
	return slug
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}