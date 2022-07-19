package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func HandleCaptcha(responseToken string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error occured loading .env. Err: %s", err)
	}

	data := map[string]string{
		"secret":   os.Getenv("SECRET_KEY"),
		"response": responseToken,
		// "remoteip": "ip-address-of-the-user",
	}
	json_data, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(
		"https://hcaptcha.com/siteverify",
		"application/x-www-form-urlencoded",
		bytes.NewBuffer(json_data),
	)

	if err != nil {
		log.Fatal(err)
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	fmt.Println(res["json"])

	if !res["success"].(bool) {
		panic(res["error-codes"].([]interface{}))
	}
}
