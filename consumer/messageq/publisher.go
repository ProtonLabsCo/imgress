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

type ChanConfirmMsgBody struct {
	Filename      string
	FileLink      string
	AfterSize     uint
	RespQueueName string
}

type ConfirmMsgBody struct {
	Filename  string
	FileLink  string
	AfterSize uint
}

type RMQPubClient struct {
	Conn    *rabbitmq.Connection
	Chan    *rabbitmq.Channel
	ConfMsg chan ChanConfirmMsgBody
	Err     chan error
}

func NewPublisher() *RMQPubClient {
	return &RMQPubClient{
		ConfMsg: make(chan ChanConfirmMsgBody),
		Err:     make(chan error),
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
		return fmt.Errorf("error in creating rabbitmq connection with %s : %s", amqpServerURL, err.Error())
	} else {
		log.Println("Consumer: succesfully connected to RabbitMQ!")
	}

	go func() {
		<-pubCl.Conn.NotifyClose(make(chan *rabbitmq.Error)) // Listen to NotifyClose
		pubCl.Err <- errors.New("connection closed")
	}()
	pubCl.Chan, err = pubCl.Conn.Channel()
	if err != nil {
		return fmt.Errorf("channel: %s", err)
	}
	return nil
}

func (pubCl *RMQPubClient) sendToQueue(confMsg ChanConfirmMsgBody) error {
	args := make(rabbitmq.Table)
	args["x-max-length"] = 5
	args["x-expires"] = 300000
	_, err := pubCl.Chan.QueueDeclare(
		confMsg.RespQueueName, // queue name
		false,                 // durable
		true,                  // auto delete
		false,                 // exclusive
		false,                 // no wait
		args,                  // arguments
	)
	if err != nil {
		log.Println(err)
		return err
	}

	rawMsgBody := ConfirmMsgBody{
		Filename:  confMsg.Filename,
		FileLink:  confMsg.FileLink,
		AfterSize: confMsg.AfterSize,
	}
	mqBody, err := json.Marshal(rawMsgBody)
	if err != nil {
		log.Println("Consumer: error encoding JSON: ", err)
		return err
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
	if err := pubCl.Chan.PublishWithContext(ctx,
		"",                    // exchange
		confMsg.RespQueueName, // queue name
		false,                 // mandatory
		false,                 // immediate
		message,               // message to publish
	); err != nil {
		log.Println(err)
		return err
	}
	log.Println("Consumer: Successfully published the confirmation message to RabbitMQ")
	return nil
}

func (pubCl *RMQPubClient) StartPublisher() {
	for {
		select {
		case err := <-pubCl.Err:
			if err != nil {
				log.Println("Consumer: connection lost, now reconnecting...: ", err)
				pubCl.Connect() // Reconnect
			}
		case msg := <-pubCl.ConfMsg:
			pubCl.sendToQueue(msg)
		}
	}
}
