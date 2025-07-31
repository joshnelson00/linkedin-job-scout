package main

import (
    "gopkg.in/gomail.v2"
    "log"
    "os"
    "time"
	"fmt"
)

func sendEvaluationsEmail() error {
    m := gomail.NewMessage()

    from := os.Getenv("EMAIL_FROM")
    to := os.Getenv("EMAIL_TO")
    smtpHost := os.Getenv("SMTP_HOST")
    smtpPort := 587 // or parse from env
    password := os.Getenv("EMAIL_PASSWORD")

    // Format current time nicely for the subject header with am/pm in local timezone
    now := time.Now().Format("Jan 1 3:04 PM MST")

    m.SetHeader("From", from)
    m.SetHeader("To", to)
    m.SetHeader("Subject", "LinkedIn Evaluations - " + now)
    m.SetBody("text/plain", "Hello,\n\nPlease find the LinkedIn Evaluations attached.\n\nThanks,\n LinkedIn Job Scout")
    m.Attach("LinkedinEvaluations.txt")

    d := gomail.NewDialer(smtpHost, smtpPort, from, password)

    if err := d.DialAndSend(m); err != nil {
        return fmt.Errorf("Could not send email: %v", err)
    }
    log.Println("Email sent successfully!")

	return nil
}
