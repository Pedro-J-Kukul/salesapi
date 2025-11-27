package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail"
)

//go:embed templates/*
var templatesFS embed.FS

// Mailer represents a mailer service.
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// New creates a new Mailer instance.
func New(host string, port int, username, password, sender string) *Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second
	return &Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send sends an email using the mailer service.
func (m *Mailer) Send(to, templateName string, data any) error {
	tmpl, err := template.ParseFS(templatesFS, "templates/"+templateName)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)                           // buffer to hold the plain text body
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data) // execute the plain body template
	if err != nil {
		return err // return error if plain body template execution fails
	}

	htmlBody := new(bytes.Buffer)                          // buffer to hold the HTML body
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data) // execute the HTML body template
	if err != nil {
		return err // return error if HTML body template execution fails
	}

	// Create a new email message
	msg := mail.NewMessage()
	msg.SetHeader("From", m.sender)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// 3 times retry logic
	for i := 0; i < 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if err == nil {
			return nil
		}
		time.Sleep(5 * time.Millisecond)
	}

	return err
}
