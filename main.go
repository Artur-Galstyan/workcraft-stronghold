package main

import (
	"log"
	"os"

	"github.com/Artur-Galstyan/workcraft-stronghold/stronghold"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var apiKey string

var (
	Version    = "dev"
	BuildDate  = "unknown"
	CommitHash = "unknown"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey = os.Getenv("WORKCRAFT_API_KEY")
	if apiKey == "" {
		log.Fatal("WORKCRAFT_API_KEY not set in environment")
	}
}

func main() {
	log.Printf("Version: %s, Build Date: %s, Commit: %s\n", Version, BuildDate, CommitHash)

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})
	stronhold := stronghold.NewStronghold(apiKey, client)
	stronhold.Run()

}
