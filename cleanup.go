package main

import (
	"fmt"
	"imgress/database"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

func CleanUp() {
	// select links from db undeleted images older than 5mins
	var images []database.Image

	for {
		last5mins := time.Now().Add(-time.Minute * 5)
		database.GDB.Where("created_at < ? AND is_deleted = false", last5mins).Find(&images)
		for _, image := range images {
			deleteFromWasabiS3(image.ImageName)
			// update db row as deleted
			database.GDB.Model(&image).Update("is_deleted", true)
		}

		// check for 5-minutes-old files in every one minute
		time.Sleep(60 * time.Second)
	}
}

func deleteFromWasabiS3(filename string) error {
	err := godotenv.Load()
	if err != nil {
		return err
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
	}

	// create a s3 client session
	s3Client := s3.New(goSession)

	// create put object input
	deleteObjectInput := &s3.DeleteObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(filename),
	}
	// get file
	_, err = s3Client.DeleteObject(deleteObjectInput)

	// print if there is an error
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}
