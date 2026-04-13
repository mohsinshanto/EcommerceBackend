package config

import (
	"log"
	"os"
)

var JwtSecret []byte

func LoadJWTSecret() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-mohsinshanto-jwt-secret-change-me"
		log.Println("WARNING: JWT_SECRET is not set. Using an insecure development secret.")
	}

	JwtSecret = []byte(secret)
}
