package auth

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// EmailSender handles email dispatching via SMTP
type EmailSender struct {
	cfg *Config
}

// NewEmailSender initializes an EmailSender
func NewEmailSender(cfg *Config) *EmailSender {
	return &EmailSender{cfg: cfg}
}

// IsConfigured returns true if SMTP credentials are fully provided
func (e *EmailSender) IsConfigured() bool {
	return e.cfg != nil &&
		strings.TrimSpace(e.cfg.SMTPEmail) != "" &&
		strings.TrimSpace(e.cfg.SMTPAppPassword) != ""
}

// SendTemporaryPassword sends a temporary password to the user's email address
func (e *EmailSender) SendTemporaryPassword(toEmail, tempPassword string) error {
	if !e.IsConfigured() {
		log.Printf("[AUTH-DEV] SMTP not configured in config.json. Temporary password for <%s> is: %s", toEmail, tempPassword)
		return nil
	}

	subject := "Your Temporary Password for Rutils"
	body := fmt.Sprintf(
		"Hello,\r\n\r\n"+
			"A password reset was requested for your account.\r\n\r\n"+
			"Your temporary password is:\r\n"+
			"  %s\r\n\r\n"+
			"Please use this temporary password to log in. You will be required to set a new permanent password immediately upon login.\r\n\r\n"+
			"If you did not request this, please ignore this email or contact support.\r\n\r\n"+
			"— Rutils Security Team",
		tempPassword,
	)

	return e.sendMail(toEmail, subject, body)
}

// SendPasswordChangedNotification sends a confirmation email when a user changes their password
func (e *EmailSender) SendPasswordChangedNotification(toEmail string) error {
	if !e.IsConfigured() {
		log.Printf("[AUTH-DEV] SMTP not configured in config.json. Password change confirmation email for <%s> simulated.", toEmail)
		return nil
	}

	subject := "Security Alert: Password Changed for Rutils Account"
	body := "Hello,\r\n\r\n" +
		"This email confirms that the password for your Rutils account was just changed.\r\n\r\n" +
		"If you made this change, you can safely disregard this email.\r\n" +
		"If you did NOT change your password, please contact support or reset your password immediately.\r\n\r\n" +
		"— Rutils Security Team"

	return e.sendMail(toEmail, subject, body)
}

// sendMail formats and delivers an email via SMTP
func (e *EmailSender) sendMail(toEmail, subject, body string) error {
	host := e.cfg.SMTPHost
	if host == "" {
		host = "smtp.gmail.com"
	}
	port := e.cfg.SMTPPort
	if port == "" {
		port = "587"
	}

	from := e.cfg.SMTPEmail
	addr := fmt.Sprintf("%s:%s", host, port)

	auth := smtp.PlainAuth("", from, e.cfg.SMTPAppPassword, host)

	message := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n\r\n"+
			"%s\r\n",
		from, toEmail, subject, body,
	)

	err := smtp.SendMail(addr, auth, from, []string{toEmail}, []byte(message))
	if err != nil {
		log.Printf("[ERROR] Failed to send email to %s via %s: %v", toEmail, addr, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
