package service

import (
	"bytes"
	"fmt"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	mail "github.com/wneessen/go-mail"
)

type emailService struct {
	host     string
	port     int
	username string
	password string
	from     string
	fromName string
}

func NewEmailService(host, port, username, password, from, fromName string) domain.EmailService {
	portInt := 465
	fmt.Sscanf(port, "%d", &portInt)
	return &emailService{
		host:     host,
		port:     portInt,
		username: username,
		password: password,
		from:     from,
		fromName: fromName,
	}
}

func (s *emailService) send(to, subject, body string) error {
	m := mail.NewMsg()
	if err := m.FromFormat(s.fromName, s.from); err != nil {
		return err
	}
	if err := m.To(to); err != nil {
		return err
	}
	m.Subject(subject)
	m.SetBodyString(mail.TypeTextHTML, body)

	c, err := mail.NewClient(s.host,
		mail.WithPort(s.port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(s.username),
		mail.WithPassword(s.password),
		mail.WithTLSPortPolicy(mail.TLSMandatory),
	)
	if err != nil {
		return err
	}
	return c.DialAndSend(m)
}

func (s *emailService) sendWithAttachment(to, subject, body, filename string, data []byte) error {
	m := mail.NewMsg()
	if err := m.FromFormat(s.fromName, s.from); err != nil {
		return err
	}
	if err := m.To(to); err != nil {
		return err
	}
	m.Subject(subject)
	m.SetBodyString(mail.TypeTextHTML, body)
	m.AttachReader(filename, bytes.NewReader(data))

	c, err := mail.NewClient(s.host,
		mail.WithPort(s.port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(s.username),
		mail.WithPassword(s.password),
		mail.WithTLSPortPolicy(mail.TLSMandatory),
	)
	if err != nil {
		return err
	}
	return c.DialAndSend(m)
}

func (s *emailService) SendWelcome(to, firstName string) error {
	subject := "Welcome to AsianMum! 🐱"
	body := fmt.Sprintf(`
<div style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px">
  <h2>Welcome to AsianMum, %s! 🎉</h2>
  <p>Thank you for joining Malaysia's premium cat litter store.</p>
  <p>We offer a wide range of high-quality cat litter products to keep your home fresh and your cats happy.</p>
  <a href="https://asianmum.my/shop" style="background:#4F46E5;color:white;padding:12px 24px;border-radius:8px;text-decoration:none;display:inline-block;margin:20px 0">
    Shop Now
  </a>
  <p style="color:#666;font-size:14px">The AsianMum Team 🐾</p>
</div>`, firstName)
	return s.send(to, subject, body)
}

func (s *emailService) SendEmailVerification(to, firstName, verifyURL string) error {
	subject := "Verify your AsianMum email address"
	body := fmt.Sprintf(`
<div style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px">
  <h2>Hi %s, please verify your email</h2>
  <p>Click the button below to verify your email address:</p>
  <a href="%s" style="background:#4F46E5;color:white;padding:12px 24px;border-radius:8px;text-decoration:none;display:inline-block;margin:20px 0">
    Verify Email
  </a>
  <p style="color:#666;font-size:12px">This link expires in 48 hours.</p>
</div>`, firstName, verifyURL)
	return s.send(to, subject, body)
}

func (s *emailService) SendPasswordReset(to, firstName, resetURL string) error {
	subject := "Reset your AsianMum password"
	body := fmt.Sprintf(`
<div style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px">
  <h2>Hi %s, reset your password</h2>
  <p>Click the button below to reset your password:</p>
  <a href="%s" style="background:#DC2626;color:white;padding:12px 24px;border-radius:8px;text-decoration:none;display:inline-block;margin:20px 0">
    Reset Password
  </a>
  <p style="color:#666;font-size:12px">This link expires in 1 hour.</p>
</div>`, firstName, resetURL)
	return s.send(to, subject, body)
}

func (s *emailService) SendOrderConfirmation(to, firstName, orderNo string, total float64) error {
	subject := fmt.Sprintf("Order Confirmed: %s", orderNo)
	body := fmt.Sprintf(`
<div style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px">
  <h2>Thank you for your order, %s! 🛍️</h2>
  <p>Your order <strong>%s</strong> has been confirmed.</p>
  <p>Order Total: <strong>MYR %.2f</strong></p>
  <a href="https://asianmum.my/orders" style="background:#4F46E5;color:white;padding:12px 24px;border-radius:8px;text-decoration:none;display:inline-block;margin:20px 0">
    View Order
  </a>
  <p style="color:#666;font-size:14px">Thank you for choosing AsianMum! 🐱</p>
</div>`, firstName, orderNo, total)
	return s.send(to, subject, body)
}

func (s *emailService) SendOrderShipped(to, firstName, orderNo, courier, trackingNo string) error {
	subject := fmt.Sprintf("Your order %s has been shipped! 📦", orderNo)
	body := fmt.Sprintf(`
<div style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px">
  <h2>Great news, %s! Your order is on its way 🚚</h2>
  <p>Order <strong>%s</strong> has been shipped.</p>
  <p><strong>Courier:</strong> %s</p>
  <p><strong>Tracking Number:</strong> %s</p>
  <p style="color:#666;font-size:14px">AsianMum Team 🐾</p>
</div>`, firstName, orderNo, courier, trackingNo)
	return s.send(to, subject, body)
}

func (s *emailService) SendReceiptWithPDF(to, firstName, orderNo string, pdfData []byte) error {
	subject := fmt.Sprintf("Your Receipt for %s 🧾", orderNo)
	body := fmt.Sprintf(`
<div style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px">
  <h2>Payment Confirmed, %s! 🎉</h2>
  <p>Your payment for order <strong>%s</strong> has been received.</p>
  <p>Please find your receipt attached to this email.</p>
  <a href="https://asianmum.my/orders" style="background:#4F46E5;color:white;padding:12px 24px;border-radius:8px;text-decoration:none;display:inline-block;margin:20px 0">
    View Order
  </a>
  <p style="color:#666;font-size:14px">Thank you for choosing AsianMum! 🐱</p>
</div>`, firstName, orderNo)
	return s.sendWithAttachment(to, subject, body, orderNo+".pdf", pdfData)
}

// ── No-op implementation ─────────────────────────────────────────────────────

type noopEmailService struct{}

func NewNoopEmailService() domain.EmailService { return &noopEmailService{} }

func (n *noopEmailService) SendWelcome(_, _ string) error                         { return nil }
func (n *noopEmailService) SendEmailVerification(_, _, _ string) error            { return nil }
func (n *noopEmailService) SendPasswordReset(_, _, _ string) error                { return nil }
func (n *noopEmailService) SendOrderConfirmation(_, _, _ string, _ float64) error { return nil }
func (n *noopEmailService) SendOrderShipped(_, _, _, _, _ string) error           { return nil }
func (n *noopEmailService) SendReceiptWithPDF(_, _, _ string, _ []byte) error     { return nil }
