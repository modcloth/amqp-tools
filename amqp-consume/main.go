package main

import (
	"flag"
	"log"
)

import (
	. "github.com/modcloth/amqp-tools"
)

var (
	uri = flag.String("U",
		"amqp://guest:guest@localhost:5672", "AMQP Connection URI")
	rabbitmqLogs = flag.Bool("rabbitmq.logs",
		false, "Consume from amq.rabbitmq.logs and amq.rabbitmq.trace")
)

func main() {
	flag.Parse()

	if *rabbitmqLogs {
		log.Println("Tailing RabbitMQ Logs")
		TailRabbitLogs(*uri)
	}
}
