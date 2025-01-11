package auth

import (
	"bytes"
	"code-garden-server/internal/database"
	"code-garden-server/internal/services/emails"
	"fmt"
	"html/template"
)

type Service struct {
	db *database.DBClient
}

func NewAuthService(db *database.DBClient) *Service {
	return &Service{
		db,
	}
}

func (as *Service) LoginEmailPassword(email, password string) {
	fmt.Println(email, password)
}

func (as *Service) LoginGithub() {

}

func (as *Service) LoginWithEmail(email string) error {
	type input struct {
		Token string
	}

	// generate token
	token := "sample token"
	var htmlBuf bytes.Buffer
	var textBuf bytes.Buffer

	// build email
	emailTemplatePath := "./internal/services/emails/templates/login.tmpl"
	emailTemplateText := "./internal/services/emails/templates/login.txt"

	tmplHtml := template.Must(template.ParseFiles(emailTemplatePath))
	tmplText := template.Must(template.ParseFiles(emailTemplateText))

	err := tmplHtml.Execute(&htmlBuf, input{token})
	if err != nil {
		return err
	}

	err = tmplText.Execute(&textBuf, input{token})
	if err != nil {
		return err
	}

	err = emails.SendMail(emails.Mail{
		Emails:  []string{email},
		Html:    htmlBuf.String(),
		Text:    textBuf.String(),
		Subject: "Sign in to Code Garden",
	})

	return err
}
