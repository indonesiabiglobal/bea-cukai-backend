package helper

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config func to get env value from key ---
func GetEnv(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Print("Error loading .env file", err)
	}
	return os.Getenv(key)
}
