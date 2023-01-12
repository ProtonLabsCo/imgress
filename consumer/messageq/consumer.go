package messageq

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	rabbitmq "github.com/rabbitmq/amqp091-go"
)

type CompressMsgBody struct {
	ImageBuffer      []byte
	ImageName        string
	CompressionLevel int
	RespQueueName    string
}

type RMQConsClient struct {
	Conn *rabbitmq.Connection
	Chan *rabbitmq.Channel
	Err  chan error
}

func NewConsumer() *RMQConsClient {
	return &RMQConsClient{
		Err: make(chan error),
	}
}

func (consCl *RMQConsClient) Connect() error {
	amqpServerURL := os.Getenv("RABBITMQ_SERVER_URL")

	var err error
	for i := 0; i < 5; i++ {
		consCl.Conn, err = rabbitmq.Dial(amqpServerURL)
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
		<-consCl.Conn.NotifyClose(make(chan *rabbitmq.Error)) // Listen to NotifyClose
		consCl.Err <- errors.New("connection closed")
	}()
	consCl.Chan, err = consCl.Conn.Channel()
	if err != nil {
		return fmt.Errorf("channel: %s", err)
	}

	_, err = consCl.Chan.QueueDeclare(
		"compress", // queue name
		true,       // durable
		false,      // auto delete
		false,      // exclusive
		false,      // no wait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("error declaring a RabbitMQ Queue: %s", err)
	}

	return nil
}
