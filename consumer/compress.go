package main

import (
	"log"
	"strings"

	"imgress-consumer/messageq"

	"github.com/google/uuid"
	"github.com/h2non/bimg"
	rabbitmq "github.com/rabbitmq/amqp091-go"
)

const (
	S3Endpoint   = "https://s3.eu-central-2.wasabisys.com"
	S3BucketName = "imgress1"
)

func ImageCompressing(messageBody messageq.CompressMsgBody, RMQChanProd *rabbitmq.Channel) {
	buffer := messageBody.ImageBuffer
	quality := messageBody.CompressionLevel
	orgFilename := messageBody.ImageName

	uuid_str := strings.Replace(uuid.New().String(), "-", "", -1)
	filename := uuid_str[len(uuid_str)-8:] + "_" + orgFilename

	compressed, err := bimg.NewImage(buffer).Process(bimg.Options{Quality: quality})
	if err != nil {
		log.Printf("Consumer: Error Compressing Image: %s", err)
		// TODO: STOP PROCESS OR SEND FAILURE MESSAGE
	}

	// TODO: ERROR HANDLING IN GOROUTINES
	// go UploadToWasabiS3(compressed, filename)

	filelink := S3Endpoint + "/" + S3BucketName + "/" + filename
	err = messageq.SendToQueue(filename, filelink, uint(len(compressed)), messageBody.QueueName, RMQChanProd)
	if err != nil {
		log.Printf("Consumer: Error Sending Confirmation: %s", err)
		// TODO: STOP PROCESS OR SEND FAILURE MESSAGE
	}
}
