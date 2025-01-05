package config

import (
	"errors"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func init() {
	log.Println("loading env")
	err := godotenv.Load()
	if err != nil {
		panic(errors.Join(err, errors.New("failed to load env")))
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
