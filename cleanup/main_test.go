package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// MockS3Client is a mock implementation of the AWS S3 API for testing.
type MockS3Client struct {
	s3iface.S3API
	DeleteObjectOutput s3.DeleteObjectOutput
	DeleteObjectErr    error
	NewFn              func(*session.Session) s3iface.S3API
}

func (m *MockS3Client) New(sess *session.Session) s3iface.S3API {
	if m.NewFn != nil {
		return m.NewFn(sess)
	}
	return m
}

func (m *MockS3Client) DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return &m.DeleteObjectOutput, m.DeleteObjectErr
}

func TestDeleteFromWasabiS3(t *testing.T) {
	// Create a new instance of the MockS3Client
	mockS3Client := &MockS3Client{
		DeleteObjectOutput: s3.DeleteObjectOutput{},
		DeleteObjectErr:    nil,
	}
	mockS3Client.NewFn = func(sess *session.Session) s3iface.S3API {
		return mockS3Client
	}

	// Call the function being tested
	err := deleteFromWasabiS3("example-filename.txt")

	// Assert that the function returned no errors
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}
}

func TestDeleteFromWasabiS3WithError(t *testing.T) {
	// Create a new instance of the MockS3Client with an error
	mockS3Client := &MockS3Client{
		DeleteObjectOutput: s3.DeleteObjectOutput{},
		DeleteObjectErr:    awserr.New("SomeErrorType", "SomeErrorMessage", errors.New("SomeInnerError")),
	}
	mockS3Client.NewFn = func(sess *session.Session) s3iface.S3API {
		return mockS3Client
	}

	// Call the function being tested
	err := deleteFromWasabiS3("example-filename.txt")

	// Assert that the function returned an error
	if err == nil {
		t.Fatalf("Expected an error, but got nil")
	}
}
