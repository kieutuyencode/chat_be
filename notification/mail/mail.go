package mail

import (
	"backend/config"
	"html/template"

	"github.com/cockroachdb/errors"
	"github.com/wneessen/go-mail"
	"go.uber.org/fx"
)

type Mail struct {
	*mail.Client
	from     string
	Template *Template
}

type mailParams struct {
	fx.In
	Env      *config.Env
	Template *Template
}

func newMail(p mailParams) (*Mail, error) {
	client, err := mail.NewClient(
		p.Env.MailHost,
		mail.WithPort(p.Env.MailPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(p.Env.MailUser),
		mail.WithPassword(p.Env.MailPassword),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create mail client")
	}

	return &Mail{
		Client:   client,
		Template: p.Template,
		from:     p.Env.MailUser,
	}, nil
}

func (m *Mail) Send(params ...*SendParams) error {
	for _, p := range params {
		msg := mail.NewMsg()
		msg.From(m.from)
		msg.To(p.To...)
		msg.Subject(p.Subject)

		if err := msg.AddAlternativeHTMLTemplate(p.Template, p.Data); err != nil {
			return errors.Wrap(err, "failed to add HTML template to mail body")
		}

		if err := m.Client.DialAndSend(msg); err != nil {
			return errors.Wrap(err, "failed to send mail")
		}
	}

	return nil
}

func (m *Mail) SendSignIn(p *SendSignInParams) error {
	return m.Send(&SendParams{
		To:       p.To,
		Subject:  SubjectSignIn,
		Template: m.Template.SignIn,
		Data: map[string]any{
			"Code":            p.Code,
			"ExpiresInMinute": p.ExpiresInMinute,
		},
	})
}

type SendParams struct {
	To       []string
	Subject  string
	Template *template.Template
	Data     any
}

type SendSignInParams struct {
	To              []string
	Code            string
	ExpiresInMinute int
}
