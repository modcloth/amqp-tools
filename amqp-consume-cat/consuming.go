package main

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
	"github.com/modcloth/amqp-tools"
	"github.com/streadway/amqp"
)

var (
	outDirFlag        = flag.String("d", "", "Output directory for messages. If not specified, output will go to stdout.")
	continuousConsume = flag.Bool("continuous", false, "If true, consume indefinitely ; otherwise, exit when queue is emptied.")
	prettyPrint       = flag.Bool("pretty", false, "Print more human-readable JSON. Should not be used if piping into another application.")
	keepMessages      = flag.Bool("keep", false, "If set to false, messages will be purged from the queue after reading. Only applies for continuous: false")
	consumerTag       string
)

// QueueBinding is used for establishing binding list
type QueueBinding struct {
	QueueName  string
	RoutingKey string
	Exchange   string
	parts      []string

	NoWait bool
	Args   amqp.Table
}

// QueueBindings is simply an array of QueueBinding structs
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

// Set is used by flag to assign contents to a custom type
func (qb *QueueBindings) Set(value string) error {
	qbparts := strings.Split(value, "/")
	if len(qbparts) != 3 {
		return errors.New("queue binding argument requires exchange, queue name, and routing key and NOTHING else")
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

// ConsumeForBindings establishes a consumer connection and passes the deliveries into a provided channel
func ConsumeForBindings(connectionUri string, bindings QueueBindings, deliveries chan interface{}, debugger amqptools.Debugger) {
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
					deliveries <- delivery
				}
			}
		}
	}
}

// HandleDelivery handles the amqp.Delivery object and either prints it out or
// writes it to a files
func HandleDelivery(delivery amqp.Delivery, debugger amqptools.Debugger) {
	addlData := make(map[string]interface{})
	addlData["BodyAsString"] = string(delivery.Body)
	deliveryPlus := &amqptools.DeliveryPlus{
		delivery,
		addlData,
	}

	var jsonBytes []byte
	var err error

	// necessary because otherwise it isn't unmarshalable
	deliveryPlus.RawDelivery.Acknowledger = nil

	if *prettyPrint {
		jsonBytes, err = json.MarshalIndent(deliveryPlus, "", "\t")
		if debugger.WithError(err, "Unable to marshal delivery into JSON.") {
			return
		}
	} else {
		jsonBytes, err = json.Marshal(deliveryPlus)
		if debugger.WithError(err, "Unable to marshal delivery into JSON.") {
			return
		}
	}

	if len(*outDirFlag) == 0 {
		fmt.Println(fmt.Sprintf("%s", string(jsonBytes)))
	} else {
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
		}
		debugger.Print(fmt.Sprintf("Data written to %s", fileName))

		err = file.Close()
		debugger.WithError(err, fmt.Sprintf("Unable to close file '%s'.", fileName))
	}

	if *keepMessages {
		err = delivery.Reject(true)
	} else {
		err = delivery.Ack(false)
	}
	if debugger.WithError(err, "Unable to Ack a message") {
		return
	}
}
