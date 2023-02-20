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
	if messageBody.ImageName == "error" {
		confMsg := messageq.ChanConfirmMsgBody{
			Filename:      "error",
			FileLink:      "not available",
			AfterSize:     0,
			RespQueueName: messageBody.RespQueueName,
		}
		pubClient.ConfMsg <- confMsg
		return
	}
	compressed := ImageCompressing(messageBody)
	if compressed == nil {
		confMsg := messageq.ChanConfirmMsgBody{
			Filename:      "error",
			FileLink:      "not available",
			AfterSize:     0,
			RespQueueName: messageBody.RespQueueName,
		}
		pubClient.ConfMsg <- confMsg
		return
	}

	uniqueFilename := messageBody.ImageName
	if err := UploadToWasabiS3(compressed, uniqueFilename); err != nil { // should it be called concurrently?
		confMsg := messageq.ChanConfirmMsgBody{
			Filename:      "error",
			FileLink:      "not available",
			AfterSize:     0,
			RespQueueName: messageBody.RespQueueName,
		}
		pubClient.ConfMsg <- confMsg
		return
	}

	filelink := S3Endpoint + "/" + S3BucketNameCompressed + "/" + uniqueFilename
	confMsg := messageq.ChanConfirmMsgBody{
		Filename:      uniqueFilename,
		FileLink:      filelink,
		AfterSize:     uint(len(compressed)),
		RespQueueName: messageBody.RespQueueName,
	}
	pubClient.ConfMsg <- confMsg
}

func main() {
	pubClient := messageq.NewPublisher()
	pubClient.Connect()
	go pubClient.StartPublisher()
	defer pubClient.Chan.Close()
	defer pubClient.Conn.Close()

	consClient := messageq.NewConsumer()
	// Start Consuming
	for {
		consClient.Connect()

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
			log.Fatal("Consumer: ", err)
		}
		log.Println("Consumer: Waiting for messages...")
		// Make a channel to receive messages into infinite loop.
		connected := true
		for connected { //receive loop
			select {
			case err := <-consClient.Err:
				if err != nil {
					log.Println("Consumer: connection lost, now reconnecting...: ", err)
				}
				connected = false
				break
			default:
				forever := make(chan bool)
				go func() {
					for message := range messages {
						rawMsgBody := messageq.CompressMsgBody{}
						err := json.Unmarshal(message.Body, &rawMsgBody)
						if err != nil {
							log.Println("Consumer: Error decoding JSON")
							rawMsgBody.ImageName = "error"
						}
						log.Println("Consumer: Recieved an image with name: ", rawMsgBody.ImageName)
						go handleConsumedMsg(rawMsgBody, pubClient)
					}
				}()
				<-forever
			}
		}
	}
	consClient.Chan.Close()
	consClient.Conn.Close()
}
