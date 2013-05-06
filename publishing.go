package amqptools

import (
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"path/filepath"
)

import (
	"github.com/streadway/amqp"
)

type PublishResult struct {
	Message string
	Error   error
	IsFatal bool
}

func PublishFiles(files chan string, connectionUri, exchange,
	routingKey string, mandatory, immediate bool, deliveryProperties DeliveryProperties, results chan *PublishResult) {

	var err error
	var body []byte
	var message *amqp.Publishing
	var messages chan *amqp.Publishing

	messages = make(chan *amqp.Publishing)

	go Publish(messages, connectionUri, exchange, routingKey, mandatory, immediate, results)

	for file := range files {
		if body, err = ioutil.ReadFile(file); err != nil {
			results <- &PublishResult{"Failed to read file " + file, err, false}
			continue
		}

		message = NewAmqpPublishingWithDelivery(deliveryProperties, body)

		if message.ContentType == "" {
			message.ContentType = mime.TypeByExtension(filepath.Ext(file))
		}

		messages <- message
	}
}

func Publish(messages chan *amqp.Publishing, connectionUri, exchange,
	routingKey string, mandatory, immediate bool, results chan *PublishResult) {

	var err error
	var conn *amqp.Connection
	var channel *amqp.Channel

	defer close(results)

	if conn, err = amqp.Dial(connectionUri); err != nil {
		results <- &PublishResult{"Failed to connect", err, true}
		return
	}

	defer conn.Close()

	if channel, err = conn.Channel(); err != nil {
		results <- &PublishResult{"Failed to get channel", err, true}
		return
	}

	pubAcks, pubNacks := channel.NotifyConfirm(make(chan uint64), make(chan uint64))
	chanClose := channel.NotifyClose(make(chan *amqp.Error))
	if err = channel.Confirm(false); err != nil {
		results <- &PublishResult{
			"Failed to put channel into confirm mode",
			err,
			true,
		}
		return
	}

	for message := range messages {
		err = channel.Publish(exchange, routingKey, mandatory, immediate, *message)
		if err != nil {
			results <- &PublishResult{"Failed to publish message", err, false}
			continue
		}

		select {
		case err = <-chanClose:
			results <- &PublishResult{"Channel closed!", err, true}
		case <-pubAcks:
			results <- &PublishResult{
				fmt.Sprintf("Published to exchange '%s' routing key '%v': %+v", exchange, routingKey, message),
				nil,
				false,
			}
		case <-pubNacks:
			results <- &PublishResult{"Received basic.nack for message", errors.New("'basic.nack'"), false}
		}
	}
}
