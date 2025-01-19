package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()

	if err != nil {
		panic(errors.Join(err, errors.New("failed to load env")))
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
