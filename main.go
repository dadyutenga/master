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
	client.Get("/billing", h.ClientBillingPage)
	client.Post("/billing/payment", h.ClientSubmitPayment)
	client.Post("/billing/receipt", h.ClientUploadReceipt)

	// Instance management
	client.Get("/instances", h.ClientInstancesList)
	client.Get("/instances/new", h.ClientShownewInstance)
	client.Post("/instances/new", h.ClientCreateInstance)
	client.Get("/instances/:id", h.ClientInstanceDetail)
	client.Post("/instances/:id/pause", h.ClientPauseInstance)
	client.Post("/instances/:id/restart", h.ClientRestartInstance)
	client.Post("/instances/:id/disable", h.ClientDisableInstance)
	client.Post("/instances/:id/pay", h.ClientPayInstance)

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
	admin.Post("/users/:userId/change-password", h.AdminChangePassword)
	admin.Post("/tenants/:id/deployments/start", h.StartTenantDeployment)
	admin.Post("/tenants/:id/deployments/stop", h.StopTenantDeployment)
	admin.Get("/tenants/:id/billing", h.TenantBillingPage)
	admin.Post("/tenants/:id/billing/charge", h.ChargeTenant)
	admin.Post("/tenants/:id/billing/payment", h.RecordPayment)
	admin.Post("/tenants/:id/billing/status", h.UpdateBillingStatus)
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

	// Instance Management
	admin.Get("/instances", h.AdminInstancesList)
	admin.Get("/instances/:id", h.AdminInstanceDetail)
	admin.Post("/instances/:id/enable", h.AdminEnableInstance)
	admin.Post("/instances/:id/disable", h.AdminDisableInstance)
	admin.Post("/instances/:id/start", h.AdminStartInstance)
	admin.Post("/instances/:id/stop", h.AdminStopInstance)
	admin.Post("/instances/:id/archive", h.AdminArchiveInstance)
	admin.Post("/instances/:id/delete", h.AdminDeleteInstance)
	admin.Post("/instances/:id/price", h.AdminUpdateInstancePrice)
	admin.Post("/instances/:id/billing", h.AdminUpdateInstanceBilling)

	// Billing Packages
	admin.Get("/billing-packages", h.AdminBillingPackages)
	admin.Post("/billing-packages/new", h.AdminCreateBillingPackage)
	admin.Get("/billing-packages/:id/edit", h.AdminEditBillingPackage)
	admin.Post("/billing-packages/:id/edit", h.AdminUpdateBillingPackage)
	admin.Post("/billing-packages/:id/delete", h.AdminDeleteBillingPackage)

	log.Fatal(app.Listen(":8080"))
}
