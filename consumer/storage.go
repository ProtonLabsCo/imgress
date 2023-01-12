package main

import (
	"bytes"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadToWasabiS3(compressedBuffer []byte, filename string) {
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
		// TODO: STOP PROCESS OR SEND FAILURE MESSAGE
	}

	// create a s3 client session
	s3Client := s3.New(goSession)

	// create put object input
	putObjectInput := &s3.PutObjectInput{
		Body:   bytes.NewReader(compressedBuffer),
		Bucket: aws.String(S3BucketName),
		Key:    aws.String(filename),
	}

	// upload file
	_, err = s3Client.PutObject(putObjectInput)
	if err != nil {
		log.Println(err)
		// TODO: STOP PROCESS OR SEND FAILURE MESSAGE
	}
}
