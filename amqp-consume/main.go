package main

import (
	"flag"
	"fmt"
	"log"
)

import (
	. "github.com/modcloth/amqp-tools"
	"github.com/streadway/amqp"
)

var (
	uri = flag.String("U",
		"amqp://guest:guest@localhost:5672", "AMQP Connection URI")
	rabbitmqLogs = flag.Bool("rabbitmq.logs",
		false, "Consume from amq.rabbitmq.logs and amq.rabbitmq.trace")
	debug = flag.Bool("debug", false, "Show debug output")
)

func main() {
	flag.Parse()

	deliveries := make(chan amqp.Delivery)

	if *rabbitmqLogs {
		if *debug {
			log.Println("Tailing RabbitMQ Logs")
		}

		go TailRabbitLogs(*uri, deliveries, *debug)

		for delivery := range deliveries {
			fmt.Printf("%s: %s", delivery.Exchange, string(delivery.Body))
		}
	}
}
