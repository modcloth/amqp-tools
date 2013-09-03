package amqptools

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

import (
	"github.com/streadway/amqp"
)

type QueueBinding struct {
	QueueName  string
	RoutingKey string
	Exchange   string
	parts      []string

	NoWait bool
	Args   amqp.Table
}

type QueueBindings []*QueueBinding

func (qb *QueueBindings) String() string {
	return fmt.Sprint(*qb)
}

func (qb *QueueBinding) String() string {
	ret := ""
	lastIndex := len(qb.parts) - 1
	for index, part := range qb.parts {
		if len(part) == 0 {
			ret += "_"
		} else {
			ret += part
		}
		if index != lastIndex {
			ret += string(os.PathSeparator)
		}
	}
	return ret
}

func (qb *QueueBindings) Set(value string) error {
	qbparts := strings.Split(value, "/")
	if len(qbparts) != 3 {
		return errors.New("Queue Binding argument requires exchange, queue name, and routing key and NOTHING else!")
	}
	newBinding := &QueueBinding{
		Exchange:   qbparts[0],
		QueueName:  qbparts[1],
		RoutingKey: qbparts[2],
		parts:      qbparts,
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
	deliveries chan amqp.Delivery, debugger *Debugger) {

	defer close(deliveries)

	conn, err := amqp.Dial(connectionUri)
	if debugger.WithError(err, "connection.open:", err) {
		return
	}

	defer conn.Close()

	channel, err := conn.Channel()
	if debugger.WithError(err, "connection.open:", err) {
		return
	}

	var consumerChannels []consumerChannel

	for _, binding := range bindings {
		queue, err := channel.QueueDeclare(binding.QueueName, false, true, false, false, nil)
		if debugger.WithError(err, "channel.queue_declare:", err) {
			return
		}

		err = channel.QueueBind(queue.Name, binding.RoutingKey, binding.Exchange, false, nil)
		if debugger.WithError(err, "channel.queuebind", err) {
			return
		}

		consumerChan, err := channel.Consume(queue.Name, "", true, false, false, false, nil)
		if debugger.WithError(err, "channel.consume", err) {
			return
		}

		consumerChannels = append(consumerChannels, consumerChannel{consumerChan, binding})
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

func TailRabbitLogs(connectionUri string, deliveries chan amqp.Delivery, debugger *Debugger) {

	defer close(deliveries)

	conn, err := amqp.Dial(connectionUri)
	if debugger.WithError(err, "connection.open:", err) {
		return
	}

	defer conn.Close()

	channel, err := conn.Channel()
	if debugger.WithError(err, "channel.open:", err) {
		return
	}

	queue, err := channel.QueueDeclare("", false, true, false, false, nil)
	if debugger.WithError(err, "channel.queue_declare:", err) {
		return
	}

	err = channel.QueueBind(queue.Name, "#", "amq.rabbitmq.log", false, nil)
	if debugger.WithError(err, "channel.queue_bind:", err) {
		return
	}

	logs, err := channel.Consume(queue.Name, "firehose-log-consumer",
		true, false, false, false, nil)
	if debugger.WithError(err, "channel.consume", err) {
		return
	}

	debugger.Print("Consuming messages from amq.rabbitmq.log")

	err = channel.QueueBind(queue.Name, "#", "amq.rabbitmq.trace", false, nil)
	if debugger.WithError(err, "channel.queue_bind:", err) {
		return
	}

	traces, err := channel.Consume(queue.Name, "firehose-trace-consumer",
		true, false, false, false, nil)
	if debugger.WithError(err, "channel.consume:", err) {
		return
	}

	debugger.Print("Consuming messages from amq.rabbitmq.trace")

	channelCloses := channel.NotifyClose(make(chan *amqp.Error))
	connCloses := conn.NotifyClose(make(chan *amqp.Error))

	for {
		select {
		case entry := <-logs:
			deliveries <- entry
		case entry := <-traces:
			deliveries <- entry
		case err := <-connCloses:
			debugger.Print("received Connection close error:", err)
			return
		case err := <-channelCloses:
			debugger.Print("received channel close error:", err)
			return
		}
	}
}
