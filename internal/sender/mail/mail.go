package mail

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/jordan-wright/email"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
)

type Mail struct {
	Name     string
	Account  string
	Password string
	SMTP     string
	port     string
}

func (m Mail) SendMessage(to string, message string) error {
	if to == "" {
		return nil
	}

	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", m.Name, m.Account)
	e.To = []string{to}
	e.Subject = "ICON Validator Info"

	// Replace newline characters with <br> for HTML
	messageWithBr := strings.ReplaceAll(message, "\n", "<br>")

	// Convert Markdown (with <br> for new lines) to HTML
	mdToHTML := markdown.ToHTML([]byte(messageWithBr), nil, nil)

	e.HTML = mdToHTML // Set the HTML body

	err := e.Send(fmt.Sprintf("%s:%s", m.SMTP, m.port), smtp.PlainAuth("", m.Account, m.Password, m.SMTP))
	if err != nil {
		return err
	}

	return nil
}

func (m Mail) GetReceiver(uid string) string {
	u, err := db.DBInstance.GetUser(uid)
	if err != nil {
		return ""
	}
	
	return *u.Email
}

func NewMail() (*Mail, error) {
	a := os.Getenv("GMAIL_ACCOUNT")
	p := os.Getenv("GMAIL_PASSWORD")
	po := os.Getenv("GMAIL_PORT")
	s := os.Getenv("GMAIL_SMTP")
	n := os.Getenv("GMAIL_NAME")

	if a == "" || p == "" || po == "" || s == "" || n == "" {
		return nil, fmt.Errorf("GMAIL_ACCOUNT, GMAIL_PASSWORD, GMAIL_PORT, GMAIL_SMTP, GMAIL_NAME must be set")
	}

	return &Mail{
		Account:  a,
		Name:     n,
		Password: p,
		SMTP:     s,
		port:     po,
	}, nil
}

func (m Mail) SendAlert(to string, v string, w string) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", m.Name, m.Account)
	e.To = []string{to}
	e.Subject = "ICON Validator Alert"

	// html message
	msg := fmt.Sprintf("<html><body><p>Validator jailed: <b>%s</b></p><p>Wallet %s not earning rewards for the ICX delegated to this validator!</p></body></html>", v, w)
	e.HTML = []byte(msg)

	err := e.Send(fmt.Sprintf("%s:%s", m.SMTP, m.port), smtp.PlainAuth("", m.Account, m.Password, m.SMTP))
	if err != nil {
		return err
	}

	return nil
}
