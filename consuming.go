package amqptools

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

import (
	"github.com/streadway/amqp"
)

type QueueBinding struct {
	QueueName  string
	RoutingKey string
	Exchange   string

	NoWait bool
	Args   amqp.Table
}

type QueueBindings []*QueueBinding

func (qb *QueueBindings) String() string {
	return fmt.Sprint(*qb)
}

func (qb *QueueBindings) Set(value string) error {
	qbparts := strings.Split(value, "/")
	if len(qbparts) != 3 {
		return errors.New("Queue Binding argument requires exchange, " +
			"queue name and routing key and NOTHING else!")
	}
	newBinding := &QueueBinding{
		Exchange:   qbparts[0],
		QueueName:  qbparts[1],
		RoutingKey: qbparts[2],
		NoWait:     false,
		Args:       nil,
	}
	*qb = append(*qb, newBinding)
	return nil
}

type consumerChannel struct {
	deliveries <-chan amqp.Delivery
	binding    *QueueBinding
}

func ConsumeForBindings(connectionUri string, bindings QueueBindings,
	deliveries chan amqp.Delivery, debug bool) {

	defer close(deliveries)

	conn, err := amqp.Dial(connectionUri)
	if err != nil {
		if debug {
			log.Println("connection.open:", err)
		}
		return
	}

	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		if debug {
			log.Println("channel.open:", err)
		}
		return
	}

	var consumerChannels []consumerChannel

	for _, binding := range bindings {
		queue, err := channel.QueueDeclare(binding.QueueName, false, true, false, false, nil)
		if err != nil {
			if debug {
				log.Println("channel.queue_declare:", err)
			}
			return
		}

		err = channel.QueueBind(queue.Name, binding.RoutingKey, binding.Exchange, false, nil)
		if err != nil {
			if debug {
				log.Println("channel.queuebind", err)
			}
			return
		}

		consumerChan, err := channel.Consume(queue.Name, "", true, false, false, false, nil)
		if err != nil {
			if debug {
				log.Println("channel.consume", err)
			}
			return
		}

		consumerChannels = append(consumerChannels,
			consumerChannel{consumerChan, binding})
	}

	for {
		for _, ch := range consumerChannels {
			select {
			case delivery := <-ch.deliveries:
				deliveries <- delivery
			}
		}
	}
}

func TailRabbitLogs(connectionUri string, deliveries chan amqp.Delivery,
	debug bool) {

	defer close(deliveries)

	conn, err := amqp.Dial(connectionUri)
	if err != nil {
		if debug {
			log.Println("connection.open:", err)
		}
		return
	}

	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		if debug {
			log.Println("channel.open:", err)
		}
		return
	}

	queue, err := channel.QueueDeclare("", false, true, false, false, nil)
	if err != nil {
		if debug {
			log.Println("channel.queue_declare:", err)
		}
		return
	}

	err = channel.QueueBind(queue.Name, "#", "amq.rabbitmq.log", false, nil)
	if err != nil {
		if debug {
			log.Println("channel.queue_bind:", err)
		}
		return
	}

	logs, err := channel.Consume(queue.Name, "firehose-log-consumer",
		true, false, false, false, nil)
	if err != nil {
		if debug {
			log.Println("channel.consume:", err)
		}
		return
	}

	if debug {
		log.Println("Consuming messages from amq.rabbitmq.log")
	}

	err = channel.QueueBind(queue.Name, "#", "amq.rabbitmq.trace", false, nil)
	if err != nil {
		if debug {
			log.Println("channel.queue_bind:", err)
		}
		return
	}

	traces, err := channel.Consume(queue.Name, "firehose-trace-consumer",
		true, false, false, false, nil)
	if err != nil {
		if debug {
			log.Println("channel.consume:", err)
		}
		return
	}

	if debug {
		log.Println("Consuming messages from amq.rabbitmq.trace")
	}

	channelCloses := channel.NotifyClose(make(chan *amqp.Error))
	connCloses := conn.NotifyClose(make(chan *amqp.Error))

	for {
		select {
		case entry := <-logs:
			deliveries <- entry
		case entry := <-traces:
			deliveries <- entry
		case err := <-connCloses:
			if debug {
				log.Println("received Connection close error:", err)
			}
			return
		case err := <-channelCloses:
			if debug {
				log.Println("received channel close error:", err)
			}
			return
		}
	}
}
