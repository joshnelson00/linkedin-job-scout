package main

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"os"
	"time"
)

func sendEvaluationsEmail() error {
	m := gomail.NewMessage()

	from := os.Getenv("EMAIL_FROM")
	to := os.Getenv("EMAIL_TO")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 587 // or parse from env
	password := os.Getenv("EMAIL_PASSWORD")

	// Format current time nicely for the subject header with am/pm in local timezone
	now := time.Now().Format("Jan 2 3:04 PM MST")

	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "LinkedIn Evaluations - "+now)
	m.SetBody("text/plain", "Hello,\n\nPlease find the LinkedIn Evaluations attached as an HTML file.\n\nThanks,\nLinkedIn Job Scout")
	m.Attach("LinkedinEvaluations.html")

	d := gomail.NewDialer(smtpHost, smtpPort, from, password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("could not send email: %v", err)
	}

	log.Println("✅ Email sent successfully!")
	return nil
}
