package models

import (
	"gorm.io/gorm"
	"math/rand"
	"strings"
)

type Snippet struct {
	gorm.Model
	Title    string `json:"title"`
	Code     string `json:"code"`
	Language string `json:"language"`
	PublicId string `json:"public_url" gorm:"unique"`
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
