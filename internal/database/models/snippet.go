package models

import (
	"code-garden-server/utils"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

type Snippet struct {
	BaseModel
	Code       string    `json:"code"`
	Language   string    `json:"language"`
	Output     string    `json:"output"`
	PublicId   string    `json:"publicId" gorm:"unique"`
	Owner      User      `json:"owner"`
	OwnerId    uuid.UUID `json:"ownerId" gorm:"not null"`
	Name       string    `json:"name"`
	Visibility string    `json:"visibility" gorm:"visibility default:private"`
	Forks      int       `json:"forks"`
}

// BeforeCreate hook
func (s *Snippet) BeforeCreate(tx *gorm.DB) error {
	err := s.BaseModel.BeforeCreate(tx)
	if err != nil {
		return err
	}

	randString, err := utils.GenerateRandomString(8)
	if err != nil {
		return err
	}

	s.PublicId = randString
	return nil
}
