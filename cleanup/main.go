package main

import (
	"log"
	"os"
	"time"

	"imgress-cleanup/database"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	database.ConnectDB()

	// select links from db undeleted images older than 5mins
	var images []database.Image

	for {
		last5mins := time.Now().Add(-time.Minute * 5)
		database.GDB.Where("created_at < ? AND is_deleted = false", last5mins).Find(&images)
		for _, image := range images {
			err := deleteFromWasabiS3(image.ImageName)
			// update db row as deleted
			if err != nil {
				log.Println(err)
			} else {
				database.GDB.Model(&image).Update("is_deleted", true)
			}
		}
		log.Printf("Cleanup: Deleted %d images", len(images))

		// check for 5-minutes-old files in every one minute
		time.Sleep(60 * time.Second)
	}
}

func deleteFromWasabiS3(filename string) error {
	s3Endpoint := os.Getenv("WASABI_S3_ENDPOINT")
	s3Region := os.Getenv("WASABI_REGION")
	s3BucketNameCompressed := os.Getenv("WASABI_COMPRESSED_BUCKET_NAME")
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
	if err != nil {
		return err
	}

	// create a s3 client session
	s3Client := s3.New(goSession)

	// create put object input
	deleteObjectInput := &s3.DeleteObjectInput{
		Bucket: aws.String(s3BucketNameCompressed),
		Key:    aws.String(filename),
	}
	// delete file
	_, err = s3Client.DeleteObject(deleteObjectInput)
	if err != nil {
		return err
	}
	return nil
}
