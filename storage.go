package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

func UploadToWasabiS3(compressedBuffer []byte, filename string) (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", err
	}

	s3Endpoint := os.Getenv("WASABI_S3_ENDPOINT")
	s3Region := os.Getenv("WASABI_REGION")
	s3BucketName := os.Getenv("WASABI_BUCKET_NAME")
	s3AccessKey := os.Getenv("WASABI_ACCESS_KEY")
	s3SecretKey := os.Getenv("WASABI_SECRET_KEY")

	// create a configuration
	s3Config := aws.Config{
		Credentials:      credentials.NewStaticCredentials(s3AccessKey, s3SecretKey, ""),
		Endpoint:         aws.String(s3Endpoint),
		Region:           aws.String(s3Region),
		S3ForcePathStyle: aws.Bool(true),
	}

	// create a session with the configuration above
	goSession, err := session.NewSessionWithOptions(session.Options{
		Config: s3Config,
	})

	// check if the session was created correctly.
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// create a s3 client session
	s3Client := s3.New(goSession)

	// create put object input
	putObjectInput := &s3.PutObjectInput{
		Body:   bytes.NewReader(compressedBuffer),
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(filename),
	}

	// upload file
	_, err = s3Client.PutObject(putObjectInput)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	linkToS3 := s3Endpoint + "/" + s3BucketName + "/" + filename
	return linkToS3, err
}
