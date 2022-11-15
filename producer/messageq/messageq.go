package messageq

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	rabbitmq "github.com/rabbitmq/amqp091-go"
)

var RMQConn *rabbitmq.Connection

type CompressMsgBody struct {
	ImageBuffer      []byte
	ImageName        string
	CompressionLevel int
	QueueName        string
}

type ConfirmMsgBody struct {
	Filename  string
	FileLink  string
	AfterSize uint
}

func ConnectMQ() {
	amqpServerURL := os.Getenv("RABBITMQ_SERVER_URL")

	var err error
	for i := 0; i < 5; i++ {
		RMQConn, err = rabbitmq.Dial(amqpServerURL)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Second)
	}
	if err != nil {
		log.Println("Producer: Error connecting to RabbitMQ!")
	} else {
		log.Println("Producer: Succesfully connected to RabbitMQ!")
	}

	RMQChan, err := RMQConn.Channel()
	if err != nil {
		log.Println("Producer: Error creating a RabbitMQ Channel!")
	}
	defer RMQChan.Close()

	_, err = RMQChan.QueueDeclare(
		"compress", // queue name
		true,       // durable
		false,      // auto delete
		false,      // exclusive
		false,      // no wait
		nil,        // arguments
	)
	if err != nil {
		log.Println("Producer: Error declaring a RabbitMQ Queue!")
	}
}

func SendToQueue(buffer []byte, quality int, orgFilename string, queueName string, RMQChan *rabbitmq.Channel) error {
	rawMsgBody := CompressMsgBody{
		ImageBuffer:      buffer,
		ImageName:        orgFilename,
		CompressionLevel: quality,
		QueueName:        queueName,
	}
	mqBody, err := json.Marshal(rawMsgBody)
	if err != nil {
		log.Println("Error encoding JSON")
	}
	// Create a message to publish.
	message := rabbitmq.Publishing{
		ContentType: "text/plain",
		Body:        mqBody,
	}

	// Attempt to publish a message to the queue.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := RMQChan.PublishWithContext(ctx,
		"",         // exchange
		"compress", // queue name
		false,      // mandatory
		false,      // immediate
		message,    // message to publish
	); err != nil {
		return err
	}
	log.Println("Producer: Successfully published the message to RabbitMQ")
	return nil
}

func WaitForConfirm(expectedLen int, queueName string, RMQConn *rabbitmq.Connection) []ConfirmMsgBody {
	RMQChan, err := RMQConn.Channel()
	if err != nil {
		log.Println("Producer: Error creating a RabbitMQ Channel!")
	}
	defer RMQChan.Close()

	_, err = RMQChan.QueueDeclare(
		queueName, // queue name
		true,      // durable
		true,      // auto delete
		false,     // exclusive
		false,     // no wait
		nil,       // arguments
	)
	if err != nil {
		log.Println("Producer: Error declaring a RabbitMQ Queue!")
	}

	messages, err := RMQChan.Consume(
		queueName, // queue name
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no local
		false,     // no wait
		nil,       // arguments
	)
	if err != nil {
		log.Println(err)
	}

	// Build a welcome message.
	log.Println("Producer: Waiting for confirmation messages...")

	count := 0
	var results []ConfirmMsgBody
	// Make a channel to receive messages into infinite loop.
	forever := make(chan bool)
	go func() {
		for message := range messages {
			rawMsgBody := ConfirmMsgBody{}
			err := json.Unmarshal(message.Body, &rawMsgBody)
			if err != nil {
				log.Println("Error decoding JSON")
				forever <- true
			}

			log.Println("Producer: Recieved Confirmation. Link: ", rawMsgBody.FileLink)
			results = append(results, rawMsgBody)
			count++
			if count == expectedLen {
				forever <- true
			}
		}
	}()
	<-forever
	log.Println("Producer: Confirmation Completed!")
	return results
}
