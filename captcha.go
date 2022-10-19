package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kataras/hcaptcha"
)

func HandleCaptcha(responseToken string) bool {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error occured loading .env. Err: %s", err)
	}

	secretKey := os.Getenv("HCAPTCHA_SECRET_KEY")

	client := hcaptcha.New(secretKey)
	hcaptchaResp := client.VerifyToken(responseToken)

	return hcaptchaResp.Success
}
