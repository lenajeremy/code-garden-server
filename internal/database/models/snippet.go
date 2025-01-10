package models

import (
	"math/rand"
	"strings"

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
func (s *Snippet) BeforeCreate(tx *gorm.DB) (err error) {
	s.PublicId = generateRandomString()
	return
}

func generateRandomString() string {
	const stringLength = 8
	digits, upperCase, lowerCase := "0123456789", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", "abcdefghijklmnopqrstuvwxyz"
	specialChars := "!@#$%^&*()_+"

	all := digits + upperCase + lowerCase + specialChars
	println(len(all))

	sBuilder := strings.Builder{}
	for i := 0; i < stringLength; i++ {
		index := rand.Intn(len(all))
		sBuilder.WriteByte(all[index])
	}

	return sBuilder.String()
}
