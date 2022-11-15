package main

import (
	"encoding/json"
	"log"

	"imgress-consumer/messageq"
)

func main() {
	RMQChanCons, RMQChanProd := messageq.ConnectMQ()
	defer RMQChanCons.Close()
	defer RMQChanProd.Close()
	defer messageq.RMQConn.Close()

	messages, err := RMQChanCons.Consume(
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
	}

	log.Println("Consumer: Waiting for messages...")

	// Make a channel to receive messages into infinite loop.
	forever := make(chan bool)
	go func() {
		for message := range messages {
			// log.Println("Consumer: Recieved a message. Processing..")
			rawMsgBody := messageq.CompressMsgBody{}
			err := json.Unmarshal(message.Body, &rawMsgBody)
			if err != nil {
				log.Println("Consumer: Error decoding JSON")
				forever <- true
				// TODO: SOMEHOW NOTIFY PRODUCER ABOUT THE ISSUE OR SEND FAILURE MESSAGE
			}
			log.Println("Consumer: Recieved an image with name: ", rawMsgBody.ImageName)
			// TODO: ERROR HANDLING IN GOROUTINES
			go ImageCompressing(rawMsgBody, RMQChanProd)
		}
	}()
	<-forever

	log.Println("Consumer: Done processing messages")
}
