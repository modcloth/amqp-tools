package amqptools

import (
	"errors"

	"github.com/streadway/amqp"
)

type SynchronousPublisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	pubAcks    chan uint64
	pubNacks   chan uint64
	chanClose  chan *amqp.Error
}

func NewSynchronousPublisher(connectionUri string) (*SynchronousPublisher, error) {
	var err error

	publisher := &SynchronousPublisher{}

	if publisher.connection, err = amqp.Dial(connectionUri); err != nil {
		return nil, err
	}

	if publisher.channel, err = publisher.connection.Channel(); err != nil {
		return nil, err
	}

	publisher.pubAcks, publisher.pubNacks = publisher.channel.NotifyConfirm(make(chan uint64), make(chan uint64))
	publisher.chanClose = publisher.channel.NotifyClose(make(chan *amqp.Error))

	if err = publisher.channel.Confirm(false); err != nil {
		return nil, err
	}

	return publisher, nil
}

func (me *SynchronousPublisher) Publish(message *amqp.Publishing, exchange, routingKey string, mandatory, immediate bool) error {
	var err error

	if err = me.channel.Publish(exchange, routingKey, mandatory, immediate, *message); err != nil {
		return err
	}

	select {
	case err = <-me.chanClose:
		return err
	case <-me.pubAcks:
		return nil
	case <-me.pubNacks:
		return errors.New("Received basic.nack")
	}

	return nil
}

func (me *SynchronousPublisher) Close() {
	me.channel.Close()
	me.connection.Close()
}
