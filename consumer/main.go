package main

import (
	"encoding/json"
	"log"
	"os"

	"imgress-consumer/messageq"
)

var S3Endpoint = os.Getenv("WASABI_S3_ENDPOINT")
var S3Region = os.Getenv("WASABI_REGION")
var S3BucketNameCompressed = os.Getenv("WASABI_COMPRESSED_BUCKET_NAME")
var S3AccessKey = os.Getenv("WASABI_ACCESS_KEY")
var S3SecretKey = os.Getenv("WASABI_SECRET_KEY")

func handleConsumedMsg(messageBody messageq.CompressMsgBody, pubClient *messageq.RMQPubClient) {
	compressed := ImageCompressing(messageBody)

	uniqueFilename := messageBody.ImageName
	UploadToWasabiS3(compressed, uniqueFilename) // should it be called concurrently?

	filelink := S3Endpoint + "/" + S3BucketNameCompressed + "/" + uniqueFilename
	confMsg := messageq.ChanConfirmMsgBody{
		Filename:      uniqueFilename,
		FileLink:      filelink,
		AfterSize:     uint(len(compressed)),
		RespQueueName: messageBody.RespQueueName,
	}
	pubClient.ConfMsg <- confMsg
}

func startConsumer(consClient *messageq.RMQConsClient, pubClient *messageq.RMQPubClient) {
	messages, err := consClient.Chan.Consume(
		"compress", // queue name
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no local
		false,      // no wait
		nil,        // arguments
	)
	if err != nil {
		log.Println("Consumer: ", err)
		return
	}

	log.Println("Consumer: Waiting for messages...")

	// Make a channel to receive messages into infinite loop.
	forever := make(chan bool)
	go func() {
		for message := range messages {
			rawMsgBody := messageq.CompressMsgBody{}
			err := json.Unmarshal(message.Body, &rawMsgBody)
			if err != nil {
				log.Println("Consumer: Error decoding JSON")
				return
				// TODO: SOMEHOW NOTIFY PRODUCER ABOUT THE ISSUE OR SEND FAILURE MESSAGE
			}
			log.Println("Consumer: Recieved an image with name: ", rawMsgBody.ImageName)
			// TODO: ERROR HANDLING IN GOROUTINES
			go handleConsumedMsg(rawMsgBody, pubClient)
		}
	}()
	<-forever

	log.Println("Consumer: Done processing messages")
	return
}

func main() {
	pubClient := messageq.NewPublisher()
	pubClient.Connect()
	go pubClient.StartPublisher()
	defer pubClient.Chan.Close()
	defer pubClient.Conn.Close()

	consClient := messageq.NewConsumer()
	consClient.Connect()
	defer consClient.Chan.Close()
	defer consClient.Conn.Close()
	startConsumer(consClient, pubClient)
	for {
		select {
		case err := <-consClient.Err:
			if err != nil {
				log.Println("Consumer: connection lost, now reconnecting...: ", err)
				consClient.Connect() // Reconnect
			}
			startConsumer(consClient, pubClient)
		default:
			// do nothing
		}
	}
}
