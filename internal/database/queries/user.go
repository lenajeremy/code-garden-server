package queries

import (
	"code-garden-server/internal/database"
	"code-garden-server/internal/database/models"
)

func GetUserFromEmail(email string, db *database.DBClient) (*models.User, error) {
	var user models.User
	tx := db.Model(models.User{}).First(&user, "email = ?", email)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &user, nil
}

func GetTokenFromString(tokenString string, db *database.DBClient) (*models.VerificationToken, error) {
	var token models.VerificationToken
	tx := db.Model(models.VerificationToken{}).Preload("User").First(&token, "token = ?", tokenString)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &token, nil
}
