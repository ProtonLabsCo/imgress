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
		errS := fmt.Errorf("error in creating a RabbitMQ connection with %s : %s", amqpServerURL, err.Error())
		log.Fatalln(errS)
	} else {
		log.Println("Consumer: succesfully connected to RabbitMQ!")
	}

	go func() {
		<-consCl.Conn.NotifyClose(make(chan *rabbitmq.Error)) // Listen to NotifyClose
		consCl.Err <- errors.New("connection closed")
	}()
	consCl.Chan, err = consCl.Conn.Channel()
	if err != nil {
		errS := fmt.Errorf("error creating a RabbitMQ channel: %s", err)
		log.Fatalln(errS)
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
		errS := fmt.Errorf("error declaring a RabbitMQ queue: %s", err)
		log.Fatalln(errS)
	}

	return nil
}
