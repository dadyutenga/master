package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	return render(c, auth.Register(auth.RegisterProps{}))
}

func (h *Handler) ShowLogin(c *fiber.Ctx) error {
	verified := c.Query("verified") == "1"
	return render(c, auth.Login(auth.LoginProps{Verified: verified, Error: ""}))
}

func (h *Handler) Register(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.FormValue("name"))
	company := strings.TrimSpace(c.FormValue("company_name"))
	email := strings.TrimSpace(c.FormValue("email"))
	phone := strings.TrimSpace(c.FormValue("phone"))
	pass := c.FormValue("password")
	confirm := c.FormValue("password_confirmation")

	if name == "" || company == "" || email == "" || pass == "" {
		return render(c, auth.Register(auth.RegisterProps{Error: "All fields are required."}))
	}
	if pass != confirm {
		return render(c, auth.Register(auth.RegisterProps{Error: "Passwords do not match."}))
	}
	if len(pass) < 8 {
		return render(c, auth.Register(auth.RegisterProps{Error: "Password must be at least 8 characters."}))
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

	q := generated.New(h.db)

	user, err := q.CreateUser(c.Context(), generated.CreateUserParams{
		Name:     name,
		Email:    email,
		Company:  company,
		Phone:    nullString(phone),
		Password: string(hash),
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return render(c, auth.Register(auth.RegisterProps{Error: "Email already registered."}))
		}
		return err
	}

	slug := generateSlug(q, c.Context(), company)
	domain := slug + "." + h.cfg.BaseDomain
	dbPass := randomHex(16)

	_, err = q.CreateTenant(c.Context(), generated.CreateTenantParams{
		UserID:      user.ID,
		CompanyName: company,
		Slug:        slug,
		Domain:      domain,
		DbName:      "hms_" + slug + "_db",
		DbUser:      "hms_" + slug + "_user",
		DbPassword:  dbPass,
	})
	if err != nil {
		return err
	}

	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	q.CreateVerifyToken(c.Context(), generated.CreateVerifyTokenParams{
		UserID: user.ID,
		Token:  token,
	})

	go h.mail.SendVerification(user.Email, user.Name, h.cfg.AppURL+"/verify/"+token)

	return c.Redirect("/verify-notice")
}

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
		ID:     tenant.ID,
		Status: generated.TenantStatusPendingApproval,
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