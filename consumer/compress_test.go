package main

import (
	"testing"

	"imgress-consumer/messageq"
)

func MockDownloadFromWasabiS3(filename string) ([]byte, error) {
	return []byte("test"), nil
}

func TestImageCompressing(t *testing.T) {
	compressed := ImageCompressing(messageq.CompressMsgBody{
		ImageName:        "test",
		CompressionLevel: 50,
		RespQueueName:    "test-queue",
	})

	if compressed == nil {
		t.Fail()
	}
}
