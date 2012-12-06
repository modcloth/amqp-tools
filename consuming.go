package amqptools

import (
	"log"
)

import (
	"github.com/streadway/amqp"
)

func TailRabbitLogs(connectionUri string, deliveries chan amqp.Delivery,
	debug bool) {

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

	defer close(deliveries)

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
