package models

import (
	"code-garden-server/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	BaseModel
	Email           string     `json:"email" gorm:"unique; not null"`
	Password        string     `json:"-" gorm:"not null"`
	FirstName       string     `json:"firstName" gorm:"first_name"`
	LastName        string     `json:"lastName" gorm:"last_name"`
	EmailVerified   bool       `json:"emailVerified" gorm:"email_verified"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt" gorm:"email_verified_at;nullable"`
}

type VerificationToken struct {
	BaseModel
	Token     string    `gorm:"unique;not null"`
	ExpiresAt time.Time `gorm:"expires_at"`
	Expired   bool
	UserID    uuid.UUID
	User      User `gorm:"foreignKey:UserID"`
}

func (vt *VerificationToken) BeforeCreate(tx *gorm.DB) error {
	err := vt.BaseModel.BeforeCreate(tx)
	if err != nil {
		return err
	}

	randStr, err := utils.GenerateRandomString(10)
	if err != nil {
		return err
	}
	vt.Token = randStr
	return nil
}

func (vt *VerificationToken) IsValid() bool {
	return vt.Expired || vt.ExpiresAt.After(time.Now())
}

func (vt *VerificationToken) Expire(db *gorm.DB) error {
	return db.Model(VerificationToken{}).Where("token = ?", vt.Token).Update("expired", true).Error
}