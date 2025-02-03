package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID      `gorm:"primaryKey;type:uuid" json:"id" redis:"id"`
	CreatedAt time.Time      `json:"createdAt" redis:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt" redis:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt" redis:"deletedAt"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	b.ID = uuid.New()
	return nil
}
