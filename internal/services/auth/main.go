package auth

import (
	"bytes"
	"code-garden-server/internal/database"
	"code-garden-server/internal/database/models"
	"code-garden-server/internal/database/queries"
	"code-garden-server/internal/services/emails"
	"errors"
	"fmt"
	"html/template"
	"time"

	"gorm.io/gorm"
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

func (as *Service) RegisterWithEmail(email, clientHost string) error {
	type input struct {
		ClientHost, Token string
	}

	var user = models.User{Email: email}
	tx := as.db.Create(&user)
	if tx.Error != nil {
		return tx.Error
	}

	token := models.VerificationToken{
		ExpiresAt: time.Now().Add(time.Minute * 10),
		UserID:    user.ID,
	}

	tx = as.db.Create(&token)
	if tx.Error != nil {
		return tx.Error
	}

	i := input{clientHost, token.Token}

	var htmlBuf bytes.Buffer
	var textBuf bytes.Buffer

	emailTemplatePath := "./internal/services/emails/templates/register.html"
	emailTemplateText := "./internal/services/emails/templates/register.txt"

	tmplHtml := template.Must(template.ParseFiles(emailTemplatePath))
	tmplText := template.Must(template.ParseFiles(emailTemplateText))

	err := tmplHtml.Execute(&htmlBuf, i)
	if err != nil {
		return err
	}

	err = tmplText.Execute(&textBuf, i)
	if err != nil {
		return err
	}

	err = emails.SendMail(emails.Mail{
		Emails:  []string{email},
		Html:    htmlBuf.String(),
		Text:    textBuf.String(),
		Subject: "Verify your Email",
	})

	return err
}

func (as *Service) LoginWithEmail(email, clientHost string) error {
	type input struct {
		ClientHost string
		Token      string
		UserName   string
	}

	user, err := queries.GetUserFromEmail(email, as.db)
	if err != nil {
		return err
	}

	// generate token
	token := models.VerificationToken{
		ExpiresAt: time.Now().Add(time.Minute * 10),
		User:      *user,
	}

	tx := as.db.Create(&token)
	if tx.Error != nil {
		return tx.Error
	}

	var htmlBuf bytes.Buffer
	var textBuf bytes.Buffer

	// build email
	emailTemplatePath := "./internal/services/emails/templates/login.tmpl"
	emailTemplateText := "./internal/services/emails/templates/login.txt"

	tmplHtml := template.Must(template.ParseFiles(emailTemplatePath))
	tmplText := template.Must(template.ParseFiles(emailTemplateText))

	err = tmplHtml.Execute(&htmlBuf, input{clientHost, token.Token, user.FirstName})
	if err != nil {
		return err
	}

	err = tmplText.Execute(&textBuf, input{clientHost, token.Token, user.FirstName})
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

func (as *Service) VerifyToken(token string) error {
	t, err := queries.GetTokenFromString(token, as.db)
	if err != nil {
		return err
	}

	if t.ExpiresAt.Before(time.Now()) || t.Expired {
		return errors.New("verification token has expired")
	}

	err = as.db.Transaction(func(db *gorm.DB) error {
		now := time.Now()
		tx := db.Model(&t.User).Updates(models.User{EmailVerified: true, EmailVerifiedAt: &now})
		if tx.Error != nil {
			return tx.Error
		}

		tx = db.Model(models.VerificationToken{}).Where("token = ?", t.Token).Update("expired", true)
		if tx.Error != nil {
			return tx.Error
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
