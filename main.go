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

	h := handlers.New(cfg, database, mail, store, eng)

	app.Get("/", h.Home)
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

	client := app.Group("/dashboard", middleware.Auth(store, database))
	client.Get("/", h.ClientDashboard)

	admin := app.Group("/admin",
		middleware.Auth(store, database),
		middleware.RequireRole("superadmin"),
	)
	admin.Get("/", h.AdminDashboard)
	admin.Get("/tenants", h.ListTenants)
	admin.Get("/tenants/:id", h.ShowTenant)
	admin.Post("/tenants/:id/approve", h.ApproveTenant)
	admin.Post("/tenants/:id/suspend", h.SuspendTenant)
	admin.Post("/tenants/:id/retry", h.RetryProvision)

	log.Fatal(app.Listen(":8080"))
}
