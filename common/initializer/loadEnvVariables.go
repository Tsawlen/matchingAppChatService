package initializer

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnvVariables(readychannel chan bool) {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	readychannel <- true
}
