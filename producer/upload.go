package main

import (
	"bytes"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var S3Endpoint = os.Getenv("WASABI_S3_ENDPOINT")
var S3Region = os.Getenv("WASABI_REGION")
var S3BucketNameRawImage = os.Getenv("WASABI_RAWIMAGE_BUCKET_NAME")
var S3AccessKey = os.Getenv("WASABI_ACCESS_KEY")
var S3SecretKey = os.Getenv("WASABI_SECRET_KEY")

func UploadToWasabiS3(uncompressedBuffer []byte, filename string) {
	// create a configuration
	s3Config := aws.Config{
		Credentials:      credentials.NewStaticCredentials(S3AccessKey, S3SecretKey, ""),
		Endpoint:         aws.String(S3Endpoint),
		Region:           aws.String(S3Region),
		S3ForcePathStyle: aws.Bool(true),
	}

	// create a session with the configuration above
	goSession, err := session.NewSessionWithOptions(session.Options{
		Config: s3Config,
	})
	if err != nil {
		log.Println(err)
		// TODO: STOP PROCESS OR SEND FAILURE MESSAGE (return 500)
	}

	// create a s3 client session
	s3Client := s3.New(goSession)

	// create put object input
	putObjectInput := &s3.PutObjectInput{
		Body:   bytes.NewReader(uncompressedBuffer),
		Bucket: aws.String(S3BucketNameRawImage),
		Key:    aws.String(filename),
	}

	// upload file
	_, err = s3Client.PutObject(putObjectInput)
	if err != nil {
		log.Println(err)
		// TODO: STOP PROCESS OR SEND FAILURE MESSAGE (return 500)
	}
}
