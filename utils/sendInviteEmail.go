package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendInviteEmail(to, msg string) error {
	from := os.Getenv("GOPROCESS_EMAIL")
	if from == "" {
		return fmt.Errorf("failed to find goprocess email")
	}

	password := os.Getenv("GOPROCESS_EMAIL_PASSWORD")
	if password == "" {
		return fmt.Errorf("failed to find goprocess email password")
	}

	fmt.Println("Sending email from: ", from)
	fmt.Println("Sending email password: ", password)

	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		return fmt.Errorf("failed to find smtp host")
	}

	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		return fmt.Errorf("failed to find smtp port")
	}

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from, []string{to}, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}
