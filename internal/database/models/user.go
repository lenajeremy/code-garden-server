package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	BaseModel
	Email     string    `json:"email" gorm:"unique; not null"`
	Password  string    `json:"password" gorm:"not null"`
	FirstName string    `json:"firstName" gorm:"first_name"`
	LastName  string    `json:"lastName" gorm:"last_name"`
}

type SafeUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (u User) Safe() SafeUser {
	return SafeUser{
		ID:        u.ID.String(),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}

func (u User) BeforeCreate(*gorm.DB) error {
	u.ID = uuid.New()
	return nil
}
