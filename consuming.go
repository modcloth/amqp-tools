package amqptools

import (
	"fmt"
	"log"
)

import (
	"github.com/streadway/amqp"
)

func TailRabbitLogs(connectionUri string) {
	conn, err := amqp.Dial(connectionUri)
	if err != nil {
		log.Println("connection.open:", err)
		return
	}

	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Println("channel.open:", err)
		return
	}

	queue, err := channel.QueueDeclare("", false, true, false, false, nil)
	if err != nil {
		log.Println("channel.queue_declare:", err)
		return
	}

	err = channel.QueueBind(queue.Name, "#", "amq.rabbitmq.log", false, nil)
	if err != nil {
		log.Println("channel.queue_bind:", err)
		return
	}

	logs, err := channel.Consume(queue.Name, "firehose-log-consumer",
		true, false, false, false, nil)
	if err != nil {
		log.Println("channel.consume:", err)
		return
	}

	log.Println("Consuming messages from amq.rabbitmq.log")

	err = channel.QueueBind(queue.Name, "#", "amq.rabbitmq.trace", false, nil)
	if err != nil {
		log.Println("channel.queue_bind:", err)
		return
	}

	traces, err := channel.Consume(queue.Name, "firehose-trace-consumer",
		true, false, false, false, nil)
	if err != nil {
		log.Println("channel.consume:", err)
		return
	}

	log.Println("Consuming messages from amq.rabbitmq.trace")

	channelCloses := channel.NotifyClose(make(chan *amqp.Error))

	connCloses := conn.NotifyClose(make(chan *amqp.Error))

	for {
		select {
		case entry := <-logs:
			log.Println("received entry from logs")
			fmt.Println(entry.Body)
		case entry := <-traces:
			log.Println("received entry from traces")
			fmt.Println(entry.Body)
		case err := <-connCloses:
			log.Println("received Connection close error")
			fmt.Println("ERROR:", err)
			return
		case err := <-channelCloses:
			log.Println("received channel close error")
			fmt.Println("ERROR:", err)
			return
		}
	}
}
