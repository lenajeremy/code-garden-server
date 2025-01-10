package auth

import "code-garden-server/internal/database"

type AuthService struct {
	db database.DBClient
}

func (as *AuthService) LoginEmailPassword(email, password string) {
	
}

func (as *AuthService) LoginGithub() {

}

func (as *AuthService) LoginWithEmail(email string) {

}