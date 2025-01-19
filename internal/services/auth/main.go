package auth

import (
	"bytes"
	"code-garden-server/config"
	"code-garden-server/internal/database"
	"code-garden-server/internal/database/models"
	"code-garden-server/internal/database/queries"
	"code-garden-server/internal/services/emails"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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

type VerificationOperation int

const (
	CreateJwtToken VerificationOperation = iota
	MarkUserEmailAsVerified
)

func (as *Service) LoginEmailPassword(email, password string) {
	fmt.Println(email, password)
}

func (as *Service) LoginGithub() {

}

func (as *Service) RegisterWithEmail(email, clientHost string) error {
	clientHost, _ = url.JoinPath(clientHost, "auth/verify-email")

	type input struct {
		ClientHost, Token string
	}

	var user = models.User{Email: email}
	err := as.db.Transaction(func(db *gorm.DB) error {
		tx := as.db.FirstOrCreate(&user, "email = ?", email)
		if tx.Error != nil {
			return tx.Error
		}

		if user.EmailVerified {
			return fmt.Errorf("an account with that email already exists, please login")
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

		fmt.Println(textBuf.String())

		err = emails.SendMail(emails.Mail{
			Emails:  []string{email},
			Html:    htmlBuf.String(),
			Text:    textBuf.String(),
			Subject: "Verify your Email",
		})

		return err
	})

	return err
}

func (as *Service) LoginWithEmail(email, clientHost string) error {
	type input struct {
		ClientHost string
		Token      string
	}

	user, err := queries.GetUserFromEmail(email, as.db)
	if err != nil {
		return err
	}

	if !user.EmailVerified {
		return as.RegisterWithEmail(email, clientHost)
	}

	// generate token
	token := models.VerificationToken{
		ExpiresAt: time.Now().Add(time.Minute * 10),
		UserID:    user.ID,
	}

	return as.db.Transaction(func(db *gorm.DB) error {
		tx := as.db.Create(&token)
		if tx.Error != nil {
			return tx.Error
		}

		var htmlBuf bytes.Buffer
		var textBuf bytes.Buffer

		// build email
		emailTemplatePath := "./internal/services/emails/templates/login.html"
		emailTemplateText := "./internal/services/emails/templates/login.txt"

		tmplHtml := template.Must(template.ParseFiles(emailTemplatePath))
		tmplText := template.Must(template.ParseFiles(emailTemplateText))

		clientHost, _ = url.JoinPath(clientHost, "auth/sign-in-with-token")

		err = tmplHtml.Execute(&htmlBuf, input{clientHost, token.Token})
		if err != nil {
			return err
		}

		err = tmplText.Execute(&textBuf, input{clientHost, token.Token})
		if err != nil {
			return err
		}

		fmt.Println(textBuf.String())

		err = emails.SendMail(emails.Mail{
			Emails:  []string{email},
			Html:    htmlBuf.String(),
			Text:    textBuf.String(),
			Subject: "Sign in to Code Garden",
		})

		return err
	})
}

func (as *Service) VerifyUserEmail(token string) (*models.VerificationToken, error) {
	t, err := queries.GetTokenFromString(token, as.db)
	if err != nil {
		return nil, err
	}

	if t.ExpiresAt.Before(time.Now()) || t.Expired {
		return nil, errors.New("verification token has expired")
	}

	err = as.db.Transaction(func(db *gorm.DB) error {
		tx := db.Model(models.VerificationToken{}).Where("token = ?", t.Token).Update("expired", true)
		if tx.Error != nil {
			return tx.Error
		}

		now := time.Now()
		tx = db.Model(&t.User).Updates(models.User{EmailVerified: true, EmailVerifiedAt: &now})
		if tx.Error != nil {
			fmt.Println(tx.Error)
			return tx.Error
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (as *Service) GenerateJwtFromToken(tokenStr string) (string, error) {
	token, err := as.VerifyUserEmail(tokenStr)
	if err != nil {
		return "", err
	}

	return generateTokenForUser(token.User)
}

func (as *Service) LoginWithPassword(email, password string) (string, error) {
	user := models.User{}
	tx := as.db.Model(models.User{}).First(&user, "email = ?", email)
	if tx.Error != nil {
		return "", tx.Error
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", err
	}

	// password matches
	return generateTokenForUser(user)
}

func (as *Service) RegisterWithPassword(email, password, clientHost string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = as.db.Create(&models.User{Email: email, Password: string(hashedPassword)}).Error
	if err != nil {
		return err
	}

	return as.RegisterWithEmail(email, clientHost)
}

type CustomJWTClaims struct {
	jwt.RegisteredClaims
	User models.User `json:"user"`
}

func AuthUser(r *http.Request) *models.User {
	if user, ok := r.Context().Value("User").(*models.User); !ok {
		panic(fmt.Errorf("attempting to get authenticated user in an unprotected route"))
	} else {
		return user
	}
}

func generateTokenForUser(user models.User) (string, error) {
	jwtSecret := []byte(config.GetEnv("JWT_SECRET"))

	claims := CustomJWTClaims{
		jwt.RegisteredClaims{
			Issuer:    "code-garden-server",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)), // tokens last a month
		},
		user,
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtTokenString, err := jwtToken.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return jwtTokenString, nil
}
