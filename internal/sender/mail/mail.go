package mail

import (
	"fmt"
	"net/smtp"
	"os"

	"github.com/jordan-wright/email"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
)

type Mail struct {
	Name string
	Account string
	Password string
	SMTP string
	port string
}

func (m Mail) SendMessage(to string, message string) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", m.Name, m.Account)
	e.To = []string{to}
	e.Subject = "ICON Validator Info"
	
	// html message
	msg := fmt.Sprintf("<html><body>%s</body></html>", message)
	e.HTML = []byte(msg)

	err := e.Send(fmt.Sprintf("%s:%s", m.SMTP,m.port), smtp.PlainAuth("", m.Account, m.Password, m.SMTP))
	if err != nil {
		return err
	}

	return nil
}

func (m Mail) GetReceiver(uid string) string {
	email := db.DBInstance.GetUserEmail(uid)
	return email
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
		Account: a,
		Name: n,
		Password: p,
		SMTP: s,
		port: po,
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

	err := e.Send(fmt.Sprintf("%s:%s", m.SMTP,m.port), smtp.PlainAuth("", m.Account, m.Password, m.SMTP))
	if err != nil {
		return err
	}

	return nil
}