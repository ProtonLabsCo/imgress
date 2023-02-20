package messageq

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	rabbitmq "github.com/rabbitmq/amqp091-go"
)

type ConfirmMsgBody struct {
	Filename  string
	FileLink  string
	AfterSize uint
}

type ConfirmExpected struct {
	ExpectedLen int
	QueueName   string
}

type RMQConsClient struct {
	Conn     *rabbitmq.Connection
	Chan     *rabbitmq.Channel
	ConfData chan ConfirmExpected
	Fanus    map[string](chan []ConfirmMsgBody)
	Err      chan error
}

func NewConsumer() *RMQConsClient {
	return &RMQConsClient{
		ConfData: make(chan ConfirmExpected),
		Fanus:    make(map[string](chan []ConfirmMsgBody)),
		Err:      make(chan error),
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
		errS := fmt.Errorf("error in creating RabbitMQ connection with %s : %s", amqpServerURL, err.Error())
		log.Fatalln(errS)
	} else {
		log.Println("Producer: succesfully connected to RabbitMQ!")
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
	return nil
}

func (consCl *RMQConsClient) Consumer() {
	for {
		select {
		case err := <-consCl.Err:
			if err != nil {
				log.Println("Producer: connection lost, now reconnecting...: ", err)
				consCl.Connect() // Reconnect
			}
		case confData := <-consCl.ConfData:
			go consCl.WaitForConfirm(confData)
		}
	}
}

func (consCl *RMQConsClient) WaitForConfirm(confData ConfirmExpected) {
	args := make(rabbitmq.Table)
	args["x-max-length"] = 5
	args["x-expires"] = 300000
	_, err := consCl.Chan.QueueDeclare(
		confData.QueueName, // queue name
		false,              // durable
		true,               // auto delete
		false,              // exclusive
		false,              // no wait
		args,               // arguments
	)
	if err != nil {
		log.Println("Producer: error declaring a RabbitMQ Queue!: ", err)
	}
	defer consCl.Chan.QueueDelete(
		confData.QueueName,
		false,
		false,
		false,
	)

	messages, err := consCl.Chan.Consume(
		confData.QueueName, // queue name
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no local
		false,              // no wait
		nil,                // arguments
	)
	if err != nil {
		log.Println("Producer: error while consuming: ", err)
	}

	// Build a welcome message.
	log.Println("Producer: waiting for confirmation messages...")

	count := 0
	var results []ConfirmMsgBody
	// Make a channel to receive messages into infinite loop.
	forever := make(chan bool)
	go func() {
		log.Println("Producer: confirmation started!")
		for message := range messages {
			rawMsgBody := ConfirmMsgBody{}
			err := json.Unmarshal(message.Body, &rawMsgBody)
			if err != nil {
				log.Println("Producer: error decoding JSON: ", err)
				forever <- true
			}

			log.Println("Producer: recieved Confirmation. Link: ", rawMsgBody.FileLink)
			results = append(results, rawMsgBody)
			count++
			if count == confData.ExpectedLen {
				forever <- true
			}
		}
	}()
	<-forever
	consCl.Fanus[confData.QueueName] <- results
}
