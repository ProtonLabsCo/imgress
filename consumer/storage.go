package main

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadToWasabiS3(compressedBuffer []byte, filename string) error {
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
		return err
	}

	// create a s3 client session
	s3Client := s3.New(goSession)

	// create put object input
	putObjectInput := &s3.PutObjectInput{
		Body:   bytes.NewReader(compressedBuffer),
		Bucket: aws.String(S3BucketNameCompressed),
		Key:    aws.String(filename),
	}

	// upload file
	_, err = s3Client.PutObject(putObjectInput)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func DownloadFromWasabiS3(filename string) ([]byte, error) {
	s3BucketNameRawImage := os.Getenv("WASABI_RAWIMAGE_BUCKET_NAME")

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
		return nil, err
	}

	// create a s3 client session
	s3Client := s3.New(goSession)

	// create put object input
	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(s3BucketNameRawImage),
		Key:    aws.String(filename),
	}
	// get file
	obj, err := s3Client.GetObject(getObjectInput)
	if err != nil {
		return nil, err
	}

	rawBuffer, err := io.ReadAll(obj.Body)
	if err != nil {
		return nil, err
	}

	// Delete downloaded image from Wasabi
	// create put object input
	deleteObjectInput := &s3.DeleteObjectInput{
		Bucket: aws.String(s3BucketNameRawImage),
		Key:    aws.String(filename),
	}
	// delete file
	_, err = s3Client.DeleteObject(deleteObjectInput)
	if err != nil {
		return nil, err
	}

	return rawBuffer, nil
}
