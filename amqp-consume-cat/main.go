package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

import (
	. "github.com/modcloth/amqp-tools"
	"github.com/streadway/amqp"
)

const (
	NOT_COOL_ZEUS = 86
	CONSUME_CAT   = `

    Ack      /\-/\
            /a a  \              _
           =\ Y  =/-~~~~~~-,____/ )
             '^--'          _____/
               \           /
               ||  |---'\  \
              (_(__|   ((__|

`
)

var queueBindings QueueBindings
var (
	uri = flag.String("U",
		"amqp://guest:guest@localhost:5672", "AMQP Connection URI")
	rabbitmqLogs = flag.Bool("rabbitmq.logs",
		false, "Consume from amq.rabbitmq.logs and amq.rabbitmq.trace")
	debug   = flag.Bool("debug", false, "Show debug output")
	showCat = flag.Bool("mrow", false, "")
)

func main() {
	flag.Var(&queueBindings, "q", "Queue bindings specified as "+
		"\"/\"-delimited strings of the form "+
		"\"exchange/queue-name/routing-key\"")
	flag.Parse()
	if *showCat {
		fmt.Println(CONSUME_CAT)
		return
	}

	deliveries := make(chan amqp.Delivery)

	if *rabbitmqLogs {
		if *debug {
			log.Println("Tailing RabbitMQ Logs")
		}

		go TailRabbitLogs(*uri, deliveries, *debug)

		for delivery := range deliveries {
			fmt.Printf("%s: %s", delivery.Exchange, string(delivery.Body))
		}
	} else if len(queueBindings) > 0 {
		go ConsumeForBindings(*uri, queueBindings, deliveries, *debug)

		for delivery := range deliveries {
			fmt.Printf("%s: %s\n", delivery.Exchange, string(delivery.Body))
		}
	} else {
		fmt.Println("ERROR: You must either consume rabbitmq logs or " +
			"define at least one exchange/queue/binding argument.")
		flag.Usage()
		os.Exit(NOT_COOL_ZEUS)
	}
}
