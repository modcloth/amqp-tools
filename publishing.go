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
}

func PublishFiles(files []string, connectionUri, defaultContentType, exchange,
	routingKey string, mandatory, immediate bool, resultChan chan *PublishFileResult) {

	var err error
	var conn *amqp.Connection
	var channel *amqp.Channel
	var message *amqp.Publishing

	if conn, err = amqp.Dial(connectionUri); err != nil {
		resultChan <- &PublishFileResult{"", "Failed to connect", err}
		return
	}

	defer conn.Close()

	if channel, err = conn.Channel(); err != nil {
		resultChan <- &PublishFileResult{"", "Failed to get channel", err}
		return
	}

	pubAcks, pubNacks := channel.NotifyConfirm(make(chan uint64), make(chan uint64))

	if err = channel.Confirm(false); err != nil {
		resultChan <- &PublishFileResult{"", "Failed to put channel into confirm mode", err}
		return
	}

	message = &amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  defaultContentType,
		Body:         make([]byte, 0),
	}

	for _, file := range files {
		if message.ContentType == "" {
			message.ContentType = mime.TypeByExtension(filepath.Ext(file))
		}

		if message.Body, err = ioutil.ReadFile(file); err != nil {
			resultChan <- &PublishFileResult{file, "Failed to read file", err}
			continue
		}

		if err = channel.Publish(exchange, routingKey, mandatory, immediate, *message); err != nil {
			resultChan <- &PublishFileResult{file, "Failed to publish file", err}
			continue
		}

		select {
		case <-pubAcks:
			resultChan <- &PublishFileResult{file, "Successfully published file", nil}
		case <-pubNacks:
			resultChan <- &PublishFileResult{file, "Received basic.nack for file", errors.New("'basic.nack'")}
		}
	}

	close(resultChan)
}
