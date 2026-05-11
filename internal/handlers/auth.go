package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/middleware"
	"github.com/dadyutenga/hms-control/internal/views/auth"
	"github.com/dadyutenga/hms-control/internal/views/client"
	"github.com/dadyutenga/hms-control/internal/views/home"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Home(c *fiber.Ctx) error {
	contact, err := h.contactDetails(c)
	if err != nil {
		return err
	}
	return render(c, home.Welcome(home.PageProps{Contact: contact}))
}

func (h *Handler) About(c *fiber.Ctx) error {
	contact, err := h.contactDetails(c)
	if err != nil {
		return err
	}
	return render(c, home.About(home.PageProps{Contact: contact}))
}

func (h *Handler) Contact(c *fiber.Ctx) error {
	contact, err := h.contactDetails(c)
	if err != nil {
		return err
	}
	return render(c, home.Contact(home.PageProps{Contact: contact}))
}

func (h *Handler) ShowRegister(c *fiber.Ctx) error {
	return render(c, auth.RegisterStep1(auth.RegisterProps{}))
}

func (h *Handler) ShowLogin(c *fiber.Ctx) error {
	verified := c.Query("verified") == "1"
	return render(c, auth.Login(auth.LoginProps{Verified: verified, Error: ""}))
}

// ── Step 1: Legal & Business Details ────────────────────────────────────────

func (h *Handler) RegisterStep1(c *fiber.Ctx) error {
	tin := strings.TrimSpace(c.FormValue("tin"))
	brela := strings.TrimSpace(c.FormValue("brela_number"))

	if tin == "" || brela == "" {
		return render(c, auth.RegisterStep1(auth.RegisterProps{Error: "Both TIN and BRELA number are required."}))
	}

	sess, _ := h.store.Get(c)
	sess.Set("reg_tin", tin)
	sess.Set("reg_brela", brela)
	if err := sess.Save(); err != nil {
		return render(c, auth.RegisterStep1(auth.RegisterProps{Error: "Session error."}))
	}
	return c.Redirect("/register/step2")
}

// ── Step 2: Hotel Information ───────────────────────────────────────────────

func (h *Handler) ShowRegisterStep2(c *fiber.Ctx) error {
	return render(c, auth.RegisterStep2(auth.RegisterProps{}))
}

func (h *Handler) RegisterStep2(c *fiber.Ctx) error {
	hotelName := strings.TrimSpace(c.FormValue("hotel_name"))
	address := strings.TrimSpace(c.FormValue("address"))
	city := strings.TrimSpace(c.FormValue("city"))
	country := strings.TrimSpace(c.FormValue("country"))

	if hotelName == "" || address == "" || city == "" || country == "" {
		return render(c, auth.RegisterStep2(auth.RegisterProps{Error: "All fields are required."}))
	}

	sess, _ := h.store.Get(c)
	sess.Set("reg_hotel_name", hotelName)
	sess.Set("reg_address", address)
	sess.Set("reg_city", city)
	sess.Set("reg_country", country)
	if err := sess.Save(); err != nil {
		return render(c, auth.RegisterStep2(auth.RegisterProps{Error: "Session error."}))
	}
	return c.Redirect("/register/step3")
}

// ── Step 3: Contact + Documents ────────────────────────────────────────────

func (h *Handler) ShowRegisterStep3(c *fiber.Ctx) error {
	sess, _ := h.store.Get(c)
	if _, ok := sess.Get("reg_tin").(string); !ok {
		return c.Redirect("/register")
	}
	return render(c, auth.RegisterStep3(auth.Step3Props{}))
}

func (h *Handler) RegisterStep3(c *fiber.Ctx) error {
	sess, _ := h.store.Get(c)

	name := strings.TrimSpace(c.FormValue("name"))
	phone := strings.TrimSpace(c.FormValue("phone"))
	email := strings.TrimSpace(c.FormValue("email"))
	pass := c.FormValue("password")
	confirm := c.FormValue("password_confirmation")

	if name == "" || email == "" || pass == "" {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Name, email, and password are required."}))
	}
	if pass != confirm {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Passwords do not match."}))
	}
	if len(pass) < 8 {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Password must be at least 8 characters."}))
	}

	q := generated.New(h.db)
	_, err := q.GetUserByEmail(c.Context(), email)
	if err == nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Email already registered."}))
	}

	brelaFile, err := c.FormFile("brela_certificate")
	if err != nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "BRELA certificate is required."}))
	}
	traFile, err := c.FormFile("tra_certificate")
	if err != nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "TRA certificate is required."}))
	}
	if brelaFile.Size > h.cfg.MaxUploadSize || traFile.Size > h.cfg.MaxUploadSize {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Files must be under 10MB each."}))
	}
	if !strings.HasPrefix(brelaFile.Header.Get("Content-Type"), "image/") ||
		!strings.HasPrefix(traFile.Header.Get("Content-Type"), "image/") {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Only image files are accepted."}))
	}

	tin, ok := sess.Get("reg_tin").(string)
	if !ok || tin == "" {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Registration session expired. Please start over."}))
	}
	brela, _ := sess.Get("reg_brela").(string)
	hotelName, _ := sess.Get("reg_hotel_name").(string)
	if hotelName == "" {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Registration session expired. Please start over."}))
	}
	address, _ := sess.Get("reg_address").(string)
	city, _ := sess.Get("reg_city").(string)
	country, _ := sess.Get("reg_country").(string)

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Registration failed."}))
	}

	var phonePtr *string
	if phone != "" {
		phonePtr = &phone
	}
	var tinPtr *string
	if tin != "" {
		tinPtr = &tin
	}
	var brelaPtr *string
	if brela != "" {
		brelaPtr = &brela
	}

	user, err := q.CreateUser(c.Context(), generated.CreateUserParams{
		Name: name, Email: email, Company: hotelName, Phone: phonePtr,
		Password: string(hash), TIN: tinPtr, BrelaNumber: brelaPtr,
	})
	if err != nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Registration failed."}))
	}
	if err := q.VerifyUser(c.Context(), user.ID); err != nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Registration failed."}))
	}

	slug := generateSlug(q, c.Context(), hotelName)
	domain := slug + "." + h.cfg.BaseDomain
	dbPass, err := randomHex(16)
	if err != nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Registration failed."}))
	}

	tenant, err := q.CreateTenant(c.Context(), generated.CreateTenantParams{
		UserID: user.ID, CompanyName: hotelName, Slug: slug, Domain: domain,
		DbName: "hms_" + slug + "_db", DbUser: "hms_" + slug + "_user", DbPassword: dbPass,
		HotelName: &hotelName, Address: &address, City: &city, Country: &country,
	})
	if err != nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Failed to create tenant."}))
	}
	if err := q.UpdateTenantStatus(c.Context(), generated.UpdateTenantStatusParams{
		ID: tenant.ID, Status: generated.TenantStatusPendingApproval,
	}); err != nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Failed to update tenant status."}))
	}

	uploadDir := filepath.Join(h.cfg.UploadDir, strconv.FormatInt(user.ID, 10))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Failed to create upload directory."}))
	}

	brelaPath, err := saveFile(brelaFile, uploadDir, "brela_")
	if err != nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Failed to save BRELA certificate."}))
	}
	if _, err := q.CreateDocument(c.Context(), generated.CreateDocumentParams{
		UserID: user.ID, DocType: "brela_certificate", Filename: brelaPath,
		OriginalName: brelaFile.Filename, MimeType: brelaFile.Header.Get("Content-Type"), SizeBytes: brelaFile.Size,
	}); err != nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Failed to record BRELA certificate."}))
	}
	traPath, err := saveFile(traFile, uploadDir, "tra_")
	if err != nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Failed to save TRA certificate."}))
	}
	if _, err := q.CreateDocument(c.Context(), generated.CreateDocumentParams{
		UserID: user.ID, DocType: "tra_certificate", Filename: traPath,
		OriginalName: traFile.Filename, MimeType: traFile.Header.Get("Content-Type"), SizeBytes: traFile.Size,
	}); err != nil {
		return render(c, auth.RegisterStep3(auth.Step3Props{Error: "Failed to record TRA certificate."}))
	}

	sess.Destroy()
	return c.Redirect("/register/success")
}

// ── Success page ────────────────────────────────────────────────────────────

func (h *Handler) ShowRegisterSuccess(c *fiber.Ctx) error {
	return render(c, auth.RegisterSuccess())
}

// ── Verify, Login, Logout ───────────────────────────────────────────────────

func (h *Handler) VerifyEmail(c *fiber.Ctx) error {
	token := c.Params("token")
	q := generated.New(h.db)
	row, err := q.GetVerifyToken(c.Context(), token)
	if err != nil {
		return c.Status(400).SendString("Invalid or expired verification link.")
	}
	if err := q.VerifyUser(c.Context(), row.Uid); err != nil {
		return c.Status(500).SendString("Failed to verify user.")
	}
	if err := q.UseVerifyToken(c.Context(), token); err != nil {
		return c.Status(500).SendString("Failed to use verification token.")
	}
	tenant, err := q.GetTenantByUserID(c.Context(), row.Uid)
	if err == nil {
		q.UpdateTenantStatus(c.Context(), generated.UpdateTenantStatusParams{
			ID: tenant.ID, Status: generated.TenantStatusPendingApproval,
		})
	}
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
	sess, _ := h.store.Get(c)
	sess.Set("userID", user.ID)
	sess.Set("role", user.Role)
	if err := sess.Save(); err != nil {
		return render(c, auth.Login(auth.LoginProps{Error: "Session error."}))
	}
	if user.Role == "admin" {
		return c.Redirect("/admin")
	}
	return c.Redirect("/dashboard")
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	sess, _ := h.store.Get(c)
	sess.Destroy()
	return c.Redirect("/login")
}

func (h *Handler) ShowChangePassword(c *fiber.Ctx) error {
	user, ok := middleware.GetUser(c)
	if !ok {
		return c.Redirect("/login")
	}
	return render(c, client.ChangePassword(client.ChangePasswordProps{User: user}))
}

func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	user, ok := middleware.GetUser(c)
	if !ok {
		return c.Redirect("/login")
	}

	current := c.FormValue("current_password")
	newPass := c.FormValue("new_password")
	confirm := c.FormValue("confirm_password")

	if newPass == "" || current == "" {
		return render(c, client.ChangePassword(client.ChangePasswordProps{User: user, Error: "All fields are required."}))
	}
	if newPass != confirm {
		return render(c, client.ChangePassword(client.ChangePasswordProps{User: user, Error: "Passwords do not match."}))
	}
	if len(newPass) < 8 {
		return render(c, client.ChangePassword(client.ChangePasswordProps{User: user, Error: "Password must be at least 8 characters."}))
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(current)); err != nil {
		return render(c, client.ChangePassword(client.ChangePasswordProps{User: user, Error: "Current password is incorrect."}))
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		return render(c, client.ChangePassword(client.ChangePasswordProps{User: user, Error: "Failed to update password."}))
	}

	q := generated.New(h.db)
	if err := q.UpdateUserPassword(c.Context(), generated.UpdateUserPasswordParams{ID: user.ID, Password: string(hash)}); err != nil {
		return render(c, client.ChangePassword(client.ChangePasswordProps{User: user, Error: "Failed to update password."}))
	}

	return render(c, client.ChangePassword(client.ChangePasswordProps{User: user, Success: true}))
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func saveFile(fh *multipart.FileHeader, dir, prefix string) (string, error) {
	ext := filepath.Ext(fh.Filename)
	randSuffix, err := randomHex(8)
	if err != nil {
		return "", err
	}
	target := filepath.Join(dir, prefix+randSuffix+ext)
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
	if base == "" {
		return "hotel"
	}
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
		if len(suffix) > 10 {
			suffix = base[:10]
		}
		slug = fmt.Sprintf("%s%d", suffix, i)
		i++
	}
	return slug
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
