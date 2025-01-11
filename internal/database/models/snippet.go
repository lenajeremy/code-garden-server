package models

import (
	"code-garden-server/utils"

	"gorm.io/gorm"
)

type Snippet struct {
	BaseModel
	Code     string `json:"code"`
	Language string `json:"language"`
	Output   string `json:"output"`
	PublicId string `json:"public_id" gorm:"unique"`
}

// BeforeCreate hook
func (s *Snippet) BeforeCreate(*gorm.DB) (err error) {
	randString, err := utils.GenerateRandomString(8)
	if err != nil {
		return err
	}

	s.PublicId = randString
	return
}
