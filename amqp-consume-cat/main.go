package main

import (
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
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

type deliveryPlus struct {
	RawDelivery  amqp.Delivery
	BodyAsString string
}

var (
	uri = flag.String("U",
		"amqp://guest:guest@localhost:5672", "AMQP Connection URI")
	rabbitmqLogs = flag.Bool("rabbitmq.logs",
		false, "Consume from amq.rabbitmq.logs and amq.rabbitmq.trace")
	debug     = flag.Bool("debug", false, "Show debug output")
	showCat   = flag.Bool("mrow", false, "")
	outputDir = flag.String("d", "", "Output directory for messages. If not specified, output will go to stdout.")
	debugger  = &Debugger{}
)

func deliver(delivery amqp.Delivery) {
	if len(*outputDir) == 0 {
		fmt.Printf("%s: %s", delivery.Exchange, string(delivery.Body))
	} else {
		deliveryPlus := &deliveryPlus{
			delivery,
			string(delivery.Body),
		}
		jsonBytes, err := json.MarshalIndent(deliveryPlus, "\t", "\t")
		if debugger.WithError(err, "Unable to marshall delivery into JSON.") {
			return
		}

		var folderName string
		if len(delivery.MessageId) > 0 {
			folderName = delivery.MessageId
		} else {
			h := sha1.New()
			fmt.Fprintf(h, "%s", jsonBytes)
			folderName = fmt.Sprintf("%x", h.Sum(nil))
		}

		var exchangeStr string
		if len(delivery.Exchange) == 0 {
			exchangeStr = "_"
		} else {
			exchangeStr = delivery.Exchange
		}

		pathParts := []string{
			strings.TrimRight(*outputDir, string(os.PathSeparator)),
			exchangeStr,
			folderName,
		}
		fullPath := strings.Join(pathParts, string(os.PathSeparator))

		fileName := fmt.Sprintf("%s%smessage.json", fullPath, string(os.PathSeparator))

		err = os.MkdirAll(fullPath, os.ModeDir|os.ModePerm)
		if debugger.WithError(err, fmt.Sprintf("Unable to create output directory '%s'.", fullPath)) {
			return
		}

		file, err := os.Create(fileName)
		if debugger.WithError(err, fmt.Sprintf("Unable to create file '%s'.", fileName)) {
			return
		}

		_, err = file.Write(jsonBytes)
		if debugger.WithError(err, fmt.Sprintf("Unable to write data into buffer for '%s'.", fileName)) {
			return
		} else {
			debugger.Print(fmt.Sprintf("Data written to %s", fileName))
		}

		err = file.Close()
		debugger.WithError(err, fmt.Sprintf("Unable to close file '%s'.", fileName))
	}
}

func main() {
	flag.Var(&queueBindings, "q", "Queue bindings specified as "+
		"\"/\"-delimited strings of the form "+
		"\"exchange/queue-name/routing-key\"")
	flag.Parse()
	if *showCat {
		fmt.Println(CONSUME_CAT)
		return
	}
	debugger.SetDebug(*debug)

	quit := make(chan bool)
	deliveries := make(chan amqp.Delivery)

	if *rabbitmqLogs || len(queueBindings) > 0 {
		if *rabbitmqLogs {
			debugger.Print("Tailing RabbitMQ Logs")

			go TailRabbitLogs(*uri, deliveries, debugger)

			go func() {
				for delivery := range deliveries {
					deliver(delivery)
				}
			}()
		}
		if len(queueBindings) > 0 {
			for _, binding := range queueBindings {
				debugger.Print(fmt.Sprintf("Binding to %s", binding))
			}
			go ConsumeForBindings(*uri, queueBindings, deliveries, debugger)

			go func() {
				for delivery := range deliveries {
					deliver(delivery)
				}
			}()
		}
	} else {
		fmt.Println("ERROR: You must either consume rabbitmq logs or " +
			"define at least one exchange/queue/binding argument.")
		flag.Usage()
		os.Exit(NOT_COOL_ZEUS)
	}
	<-quit
}
