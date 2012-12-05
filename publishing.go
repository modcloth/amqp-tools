package amqptools

import (
	"errors"
	"io/ioutil"
	"mime"
	"path/filepath"
	"time"
)

import (
	"github.com/streadway/amqp"
)

type PublishFileResult struct {
	Filename string
	Message  string
	Error    error
	IsFatal  bool
}

func PublishFiles(files chan string, connectionUri, defaultContentType, exchange,
	routingKey string, mandatory, immediate bool, results chan *PublishFileResult) {

	var err error
	var conn *amqp.Connection
	var channel *amqp.Channel
	var message *amqp.Publishing

	defer close(results)

	if conn, err = amqp.Dial(connectionUri); err != nil {
		results <- &PublishFileResult{"", "Failed to connect", err, true}
		return
	}

	defer conn.Close()

	if channel, err = conn.Channel(); err != nil {
		results <- &PublishFileResult{"", "Failed to get channel", err, true}
		return
	}

	pubAcks, pubNacks := channel.NotifyConfirm(make(chan uint64), make(chan uint64))

	if err = channel.Confirm(false); err != nil {
		results <- &PublishFileResult{"", "Failed to put channel into confirm mode", err, true}
		return
	}

	message = &amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  defaultContentType,
		Body:         make([]byte, 0),
	}

	for file := range files {
		if message.ContentType == "" {
			message.ContentType = mime.TypeByExtension(filepath.Ext(file))
		}

		if message.Body, err = ioutil.ReadFile(file); err != nil {
			results <- &PublishFileResult{
				file,
				"Failed to read file",
				err,
				false,
			}
			continue
		}

		if err = channel.Publish(exchange, routingKey, mandatory, immediate, *message); err != nil {
			results <- &PublishFileResult{
				file,
				"Failed to publish file",
				err,
				false,
			}
			continue
		}

		select {
		case <-pubAcks:
			results <- &PublishFileResult{
				file,
				"Successfully published file",
				nil,
				false,
			}
		case <-pubNacks:
			results <- &PublishFileResult{
				file,
				"Received basic.nack for file",
				errors.New("'basic.nack'"),
				false,
			}
		}
	}
}
