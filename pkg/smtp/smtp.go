package smtp

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"
)

func SendEmail(body string) error {
	subject := "PR Report Summary - " + time.Now().Format("02 Jan 2006")
	from := "vishnuk@smartrabbitz.com"
	// App-specific password Token from .env
	password := os.Getenv("EMAIL_PASSWORD")
	if password == "" {
		return fmt.Errorf("GITHUB_TOKEN not set in environment")
	}
	to := []string{"anandhu.rajan@smartrabbitz.com", "jeeva.venkatesan@smartrabbitz.com"}

	// SMTP config
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Compose message
	msg := "From: " + from + "\n" +
		"To: " + strings.Join(to, ",") + "\n" +
		"Subject: " + subject + "\n\n" + body

	// Auth
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, []byte(msg))
	return err
}
