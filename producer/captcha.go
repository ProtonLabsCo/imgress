package main

import (
	"os"

	"github.com/kataras/hcaptcha"
)

func HandleCaptcha(responseToken string) bool {
	secretKey := os.Getenv("HCAPTCHA_SECRET_KEY")

	client := hcaptcha.New(secretKey)
	hcaptchaResp := client.VerifyToken(responseToken)

	return hcaptchaResp.Success
}
