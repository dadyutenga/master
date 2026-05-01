package mailer

import (
	"fmt"
	"strconv"

	"github.com/dadyutenga/hms-control/internal/config"
	"gopkg.in/gomail.v2"
)

type Mailer struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) send(to, subject, body string) error {
	port, _ := strconv.Atoi(m.cfg.SMTPPort)
	d := gomail.NewDialer(m.cfg.SMTPHost, port, m.cfg.SMTPUser, m.cfg.SMTPPass)

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.cfg.SMTPFrom)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	return d.DialAndSend(msg)
}

func (m *Mailer) SendVerification(to, name, verifyURL string) error {
	body := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>Thanks for registering on HMS Platform.</p>
		<p>Please verify your email address:</p>
		<a href="%s" style="background:#6366f1;color:white;padding:12px 24px;
		   border-radius:6px;text-decoration:none;display:inline-block;">
			Verify Email
		</a>
		<p>This link expires in 24 hours.</p>
	`, name, verifyURL)

	return m.send(to, "Verify Your Email — HMS Platform", body)
}

func (m *Mailer) SendTenantReady(to, name, hmsURL string) error {
	body := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>Your Hotel Management System is live!</p>
		<a href="%s" style="background:#22c55e;color:white;padding:12px 24px;
		   border-radius:6px;text-decoration:none;display:inline-block;">
			Access Your HMS
		</a>
		<p>Login with the credentials you registered with.</p>
		<p>Need help? Email support@hms.co.tz</p>
	`, name, hmsURL)

	return m.send(to, "Your HMS is Ready!", body)
}