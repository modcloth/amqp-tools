package amqptools

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

import (
	"github.com/streadway/amqp"
)

var (
	outDirFlag        = flag.String("d", "", "Output directory for messages. If not specified, output will go to stdout.")
	continuousConsume = flag.Bool("continuous", false, "If true, consume indefinitely ; otherwise, exit when queue is emptied.")
	consumerTag       string
)

/*
ESTABLISHING BINDING LIST
*/

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

/*
ESTABLISHING CONNECTIONS
*/

type consumerChannel struct {
	deliveries <-chan amqp.Delivery
	binding    *QueueBinding
}

func ConsumeForBindings(connectionUri string, bindings QueueBindings, deliveries chan interface{}, debugger Debugger) {
	defer close(deliveries)

	// establish connection
	conn, err := amqp.Dial(connectionUri)
	if debugger.WithError(err, "connection.establish: ", err) {
		os.Exit(1)
	}
	defer conn.Close()
	debugger.Print("Connection established")

	// open channel
	channel, err := conn.Channel()
	if debugger.WithError(err, "channel.open: ", err) {
		os.Exit(2)
	}
	debugger.Print("Channel opened")

	// set qos
	qosValue := 10
	err = channel.Qos(qosValue, 0, false)
	if debugger.WithError(err, "channel.qos: ", err) {
		os.Exit(3)
	}
	debugger.Print(fmt.Sprintf("Channel QOS set to %d", qosValue))

	var consumerChannels []consumerChannel

	uuidBytes, err := exec.Command("uuidgen").Output()
	if debugger.WithError(err, "uuidgen", err) {
		os.Exit(4)
	}

	consumerTag = string(uuidBytes)

	for _, binding := range bindings {
		err = channel.QueueBind(binding.QueueName, binding.RoutingKey, binding.Exchange, false, nil)
		if debugger.WithError(err, "channel.queuebind ", err) {
			return
		}

		if *continuousConsume {
			//handle acking after writing file
			/*
				autoAck = false (must manually Ack)
				exclusive = true (so we only try to read from one consumer at a time)
				noLocal = true (so that if this consumer republishes to the failure queue, it will not pick the messages back up)
				noWait = true
			*/
			consumerChan, err := channel.Consume(binding.QueueName, consumerTag, false, true, true, true, nil)
			if debugger.WithError(err, "channel.consume ", err) {
				return
			}

			consumerChannels = append(consumerChannels, consumerChannel{consumerChan, binding})
		} else {
			consumerChannels = append(consumerChannels, consumerChannel{nil, binding})
		}
	}

	if *continuousConsume {
		for {
			for _, ch := range consumerChannels {
				select {
				case delivery := <-ch.deliveries:
					deliveries <- delivery
				}
			}
		}
	} else {
		count := 0
		for _, ch := range consumerChannels {
			binding := ch.binding
			debugger.Print("Getting messages from queue", binding.QueueName)
			for {
				delivery, ok, err := channel.Get(binding.QueueName, false)
				if debugger.WithError(err, "channel.get ", err) {
					os.Exit(5)
				}
				if ok == false {
					deliveries <- nil
				} else {
					count++
					fmt.Printf("count: %d\n", count)
					deliveries <- delivery
				}
			}
		}
	}
}

/*
HANDLING MESSAGES
*/

type deliveryPlus struct {
	RawDelivery  amqp.Delivery
	BodyAsString string
}

func HandleDelivery(delivery amqp.Delivery, debugger Debugger) {
	if len(*outDirFlag) == 0 {
		fmt.Println("got here")
		//fmt.Printf("%s: %s", delivery.Exchange, string(delivery.Body))
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
			strings.TrimRight(*outDirFlag, string(os.PathSeparator)),
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
	delivery.Ack(false)
}
