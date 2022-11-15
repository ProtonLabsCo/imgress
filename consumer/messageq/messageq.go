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

func ConnectMQ() (*rabbitmq.Channel, *rabbitmq.Channel) {
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
		log.Println("Consumer: Error connecting to RabbitMQ!")
	} else {
		log.Println("Consumer: Succesfully connected to RabbitMQ!")
	}

	RMQChanCons, err := RMQConn.Channel()
	if err != nil {
		log.Println("Consumer: Error creating a RabbitMQ Channel!")
	}
	_, err = RMQChanCons.QueueDeclare(
		"compress", // queue name
		true,       // durable
		false,      // auto delete
		false,      // exclusive
		false,      // no wait
		nil,        // arguments
	)
	if err != nil {
		log.Println("Consumer: Error declaring a RabbitMQ Queue!")
	}

	RMQChanProd, err := RMQConn.Channel()
	if err != nil {
		log.Println("Consumer: Error creating a RabbitMQ Channel!")
	}

	return RMQChanCons, RMQChanProd
}

func SendToQueue(filename string, filelink string, aftersize uint, queueName string, RMQChan *rabbitmq.Channel) error {
	_, err := RMQChan.QueueDeclare(
		queueName, // queue name
		true,      // durable
		true,      // auto delete
		false,     // exclusive
		false,     // no wait
		nil,       // arguments
	)
	if err != nil {
		log.Println(err)
	}

	rawMsgBody := ConfirmMsgBody{
		Filename:  filename,
		FileLink:  filelink,
		AfterSize: aftersize,
	}
	mqBody, err := json.Marshal(rawMsgBody)
	if err != nil {
		log.Println("Consumer: Error encoding JSON")
		// TODO: STOP THE PROCESS OR SEND FAILURE MESSAGE
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
		"",        // exchange
		queueName, // queue name
		false,     // mandatory
		false,     // immediate
		message,   // message to publish
	); err != nil {
		return err
	}
	log.Println("Consumer: Successfully published the confirmation message to RabbitMQ")
	return nil
}
