package mail

import (
	"html/template"
	"path"

	"github.com/cockroachdb/errors"
)

type Template struct {
	SignIn *template.Template
}

func newTemplate() (*Template, error) {
	folderPath := "notification/mail/templates"

	signInTemplate, err := template.ParseFiles(path.Join(folderPath, "sign_in.html"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse sign in template")
	}

	return &Template{
		SignIn: signInTemplate,
	}, nil
}
