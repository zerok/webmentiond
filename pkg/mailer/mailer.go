package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

// Message as it is stored by the DummyMailer when using its SendMail
// method.
type Message struct {
	From    string
	To      []string
	Subject string
	Body    string
}

// Mailer is something that can send out an email.
type Mailer interface {
	SendMail(ctx context.Context, from string, to []string, subject string, body string) error
}

// DummyMailer is an implementation of Mailer that only adds the
// passed message into the internal Messages field.
type DummyMailer struct {
	Messages []Message
}

// NewDummy creates a new DummyMailer.
func NewDummy() *DummyMailer {
	return &DummyMailer{
		Messages: make([]Message, 0, 10),
	}
}

func (m *DummyMailer) SendMail(ctx context.Context, from string, to []string, subject string, body string) error {
	m.Messages = append(m.Messages, Message{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
	})
	return nil
}

type DefaultMailer struct {
	addr      string
	auth      smtp.Auth
	tlsConfig *tls.Config
}

func New(addr string, auth smtp.Auth, tlsConfig *tls.Config) *DefaultMailer {
	m := &DefaultMailer{
		addr:      addr,
		auth:      auth,
		tlsConfig: tlsConfig,
	}
	return m
}

func (m *DefaultMailer) SendMail(ctx context.Context, from string, to []string, subject string, body string) error {
	msg := email.NewEmail()
	msg.To = to
	msg.From = from
	msg.Subject = subject
	msg.Text = []byte(body)
	err := msg.SendWithTLS(m.addr, m.auth, m.tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
