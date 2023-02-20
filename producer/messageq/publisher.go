package messageq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	rabbitmq "github.com/rabbitmq/amqp091-go"
)

type CompressMsgBody struct {
	ImageName        string
	CompressionLevel int
	RespQueueName    string
}

type RMQPubClient struct {
	Conn *rabbitmq.Connection
	Chan *rabbitmq.Channel
	Msg  chan CompressMsgBody
	Err  chan error
}

func NewPublisher() *RMQPubClient {
	return &RMQPubClient{
		Msg: make(chan CompressMsgBody),
		Err: make(chan error),
	}
}

func (pubCl *RMQPubClient) Connect() error {
	amqpServerURL := os.Getenv("RABBITMQ_SERVER_URL")

	var err error
	for i := 0; i < 5; i++ {
		pubCl.Conn, err = rabbitmq.Dial(amqpServerURL)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Second)
	}
	if err != nil {
		errS := fmt.Errorf("error in creating RabbitMQ connection with %s : %s", amqpServerURL, err.Error())
		log.Fatalln(errS)
	} else {
		log.Println("Producer: succesfully connected to RabbitMQ!")
	}

	go func() {
		<-pubCl.Conn.NotifyClose(make(chan *rabbitmq.Error)) // Listen to NotifyClose
		pubCl.Err <- errors.New("connection closed")
	}()
	pubCl.Chan, err = pubCl.Conn.Channel()
	if err != nil {
		errS := fmt.Errorf("error creating a RabbitMQ channel: %s", err)
		log.Fatalln(errS)
	}

	_, err = pubCl.Chan.QueueDeclare(
		"compress", // queue name
		true,       // durable
		false,      // auto delete
		false,      // exclusive
		false,      // no wait
		nil,        // arguments
	)
	if err != nil {
		errS := fmt.Errorf("error declaring a RabbitMQ Queue: %s", err)
		log.Fatalln(errS)
	}

	return nil
}

func (pubCl *RMQPubClient) Publisher() {
	for {
		select {
		case err := <-pubCl.Err:
			if err != nil {
				log.Println("Producer: connection lost, now reconnecting...: ", err)
				pubCl.Connect() // Reconnect
			}
		case msg := <-pubCl.Msg:
			go pubCl.SendToQueue(msg)
		}
	}
}

func (pubCl *RMQPubClient) SendToQueue(rawMsgBody CompressMsgBody) {
	mqBody, err := json.Marshal(rawMsgBody)
	if err != nil {
		log.Println("Producer: error encoding JSON: ", err)
	}
	// Create a message to publish.
	message := rabbitmq.Publishing{
		ContentType: "text/plain",
		Body:        mqBody,
	}

	// Attempt to publish a message to the queue.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pubCl.Chan.PublishWithContext(ctx,
		"",         // exchange
		"compress", // queue name
		false,      // mandatory
		false,      // immediate
		message,    // message to publish
	); err != nil {
		log.Println("Producer: error publishing to RabbitMQ: ", err)
	}
	log.Println("Producer: successfully published the message to RabbitMQ")
}
