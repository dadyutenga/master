package main

import (
	"log"

	"github.com/dadyutenga/hms-control/internal/config"
	"github.com/dadyutenga/hms-control/internal/db"
	"github.com/dadyutenga/hms-control/internal/handlers"
	"github.com/dadyutenga/hms-control/internal/mailer"
	"github.com/dadyutenga/hms-control/internal/middleware"
	"github.com/dadyutenga/hms-control/internal/provisioner"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func main() {
	cfg := config.Load()
	database := db.Connect(cfg.DBPath)
	defer database.Close()
	mail := mailer.New(cfg)

	eng := provisioner.NewEngine(cfg, database, mail)
	eng.Start()

	store := session.New(session.Config{
		KeyLookup:    "cookie:hms_session",
		CookieSecure: cfg.CookieSecure,
	})

	app := fiber.New(fiber.Config{
		ErrorHandler: handlers.ErrorHandler,
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Static("/static", "./static")
	app.Static("/img", "./public/img")
	app.Static("/uploads", "./uploads")

	h := handlers.New(cfg, database, mail, store, eng)

	app.Get("/", h.Home)
	app.Get("/about", h.About)
	app.Get("/contact", h.Contact)
	app.Get("/register", h.ShowRegister)
	app.Post("/register/step1", h.RegisterStep1)
	app.Get("/register/step2", h.ShowRegisterStep2)
	app.Post("/register/step2", h.RegisterStep2)
	app.Get("/register/step3", h.ShowRegisterStep3)
	app.Post("/register/step3", h.RegisterStep3)
	app.Get("/register/success", h.ShowRegisterSuccess)
	app.Get("/verify/:token", h.VerifyEmail)
	app.Get("/login", h.ShowLogin)
	app.Post("/login", h.Login)
	app.Post("/logout", h.Logout)
	app.Get("/verify-notice", func(c *fiber.Ctx) error {
		return c.SendString("Check your email for a verification link.")
	})

	// Stop impersonation (accessible while acting as tenant)
	app.Post("/admin/impersonate/stop", middleware.Auth(store, database), h.StopImpersonation)

	client := app.Group("/dashboard", middleware.Auth(store, database))
	client.Get("/", h.ClientDashboard)

	billingProtected := client.Group("", middleware.RequireBilling(database))
	billingProtected.Get("/details", h.ShowTenantDetails)
	billingProtected.Post("/details", h.UpdateTenantDetails)
	billingProtected.Get("/change-password", h.ShowChangePassword)
	billingProtected.Post("/change-password", h.ChangePassword)

	admin := app.Group("/admin",
		middleware.Auth(store, database),
		middleware.RequireRole("admin"),
	)
	admin.Get("/", h.AdminDashboard)
	admin.Get("/audit", h.AuditLog)
	admin.Get("/audit/export", h.ExportAuditCSV)
	// /tenants/export MUST come before /tenants/:id
	admin.Get("/tenants/export", h.ExportTenantsCSV)
	admin.Get("/tenants", h.ListTenants)
	admin.Get("/verification", h.ListVerificationTenants)
	admin.Get("/tenants/:id", h.ShowTenant)
	admin.Get("/verification/:id", h.ShowVerificationTenant)
	admin.Get("/tenants/:id/deployments/:deploymentId", h.ShowDeployment)
	admin.Get("/tenants/:id/health", h.TenantHealthCheck)
	admin.Get("/tenants/:id/logs/stream", h.StreamProvisionLogs)
	admin.Post("/tenants/:id/approve", h.ApproveTenant)
	admin.Post("/verification/:id/verify", h.VerifyTenant)
	admin.Post("/tenants/:id/suspend", h.SuspendTenant)
	admin.Post("/tenants/:id/retry", h.RetryProvision)
	admin.Post("/tenants/:id/deployments/start", h.StartTenantDeployment)
	admin.Post("/tenants/:id/deployments/stop", h.StopTenantDeployment)
	admin.Post("/tenants/:id/billing", h.UpdateTenantBilling)
	admin.Post("/tenants/:id/impersonate", h.ImpersonateTenant)
	admin.Get("/settings/contact", h.AdminContactSettings)
	admin.Post("/settings/contact", h.UpdateContactSettings)
	admin.Get("/settings/smtp", h.AdminSMTPSettings)
	admin.Post("/settings/smtp", h.UpdateSMTPSettings)
	admin.Post("/settings/smtp/test", h.TestSMTP)
	admin.Get("/settings/provisioner", h.AdminProvisionerSettings)
	admin.Post("/settings/provisioner", h.UpdateProvisionerSettings)

	// Docker Templates
	admin.Get("/docker-templates", h.ListDockerTemplates)
	admin.Get("/docker-templates/new", h.ShowCreateDockerTemplate)
	admin.Post("/docker-templates/new", h.CreateDockerTemplate)
	admin.Get("/docker-templates/:id/edit", h.ShowEditDockerTemplate)
	admin.Post("/docker-templates/:id/edit", h.UpdateDockerTemplate)
	admin.Post("/docker-templates/:id/delete", h.DeleteDockerTemplate)

	log.Fatal(app.Listen(":8080"))
}
