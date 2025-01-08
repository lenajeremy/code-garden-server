package database

import (
	"code-garden-server/config"
	"code-garden-server/internal/database/models"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBClient struct {
	*gorm.DB
}

func NewDBClient() (*DBClient, error) {
	dsn := config.GetEnv("DB_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	fmt.Println("Connection to database established")

	return &DBClient{db}, nil
}

func (db *DBClient) Setup() error {
	err := db.AutoMigrate(models.Snippet{})
	fmt.Println("making migrations")
	if err != nil {
		return err
	}
	return nil
}
