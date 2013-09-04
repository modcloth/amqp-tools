package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

import (
	. "github.com/modcloth/amqp-tools"
	"github.com/streadway/amqp"
)

const timeFormat = "Mon Jan 02 15:04:05 MST 2006"

//TODO: handle consume cat variables that leaked into amqptools

var (
	uriFlag     = flag.String("U", "amqp://guest:guest@localhost:5672", "AMQP Connection URI")
	versionFlag = flag.Bool("version", false, "Print version and exit")
	revFlag     = flag.Bool("rev", false, "Print git revision and exit")
	usageString = `Usage: %s [options] <file> [file file ...]

Parses messages consumed from RabbitMQ, extracts data and metadata of the
original message, and republishes the message. If there is only a single
filename entry and it is "-", it is assumed that the message data will be read
from stdin with entries delimited by line feeds ("\n"). Each line must be a
json-marshaled DeliveryPlus struct.  This is the output format of
amqp-consume-cat, so data may be piped from amqp-consume-cat directly into
amqp-replay-ninja.  If files are specified, the files must be valid json but
their contents may be pretty printed.

`

	debugger Debugger
)

type ErrorMessage struct {
	OriginalMessage OriginalMessage `json:"original_message"`
	OtherData       map[string]interface{}
}

type OriginalMessage struct {
	Payload    string     `json:"payload"`
	Properties Properties `json:"properties"`
	RoutingKey string     `json:"routing_key"`
	Exchange   string     `json:"exchange"`
}

type Properties struct {
	AppId           string `json:"app_id"`
	ContentType     string `json:"content_type"`
	ContentEncoding string `json:"content_encoding"`
	CorrelationId   string `json:"correlation_id"`
	MessageId       string `json:"message_id"`
	DeliveryMode    int    `json:"delivery_mode"`
	Expiration      interface{}
	Headers         interface{}
	Priority        interface{}
	Timestamp       string `json:"timestamp"`
	Type            interface{}
	UserId          interface{}
}

func init() {
	flag.Var(&debugger, "debug", "Show debug output")
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageString, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if *versionFlag {
		progName := path.Base(os.Args[0])
		if VersionString == "" {
			VersionString = "<unknown>"
		}
		fmt.Printf("%s %s\n", progName, VersionString)
		os.Exit(0)
	}

	if *revFlag {
		if RevString == "" {
			RevString = "<unknown>"
		}
		fmt.Println(RevString)
		os.Exit(0)
	}

	files := flag.Args()

	if len(files) == 0 {
		flag.Usage()
		os.Exit(5)
	}

	var conn *amqp.Connection
	var channel *amqp.Channel
	var err error

	conn, err = amqp.Dial(*uriFlag)
	debugger.Print(fmt.Sprintf("uri: %s\n", *uriFlag))
	if debugger.WithError(err, "Failed to connect ", err) {
		os.Exit(2)
	}
	debugger.Print("connection made")

	defer conn.Close()

	channel, err = conn.Channel()
	if debugger.WithError(err, "Failed to open channel ", err) {
		os.Exit(3)
	}
	debugger.Print("channel.established")

	var bytes []byte

	if len(files) == 1 && files[0] == "-" {
		stdin := bufio.NewReader(os.Stdin)
		for {
			myBytes, err := stdin.ReadString('\n')
			bytes = []byte(myBytes)
			if err != nil {
				if err != io.EOF {
					debugger.Print("ERROR:", err)
				}
				break
			}
			handleMessageBytes(bytes, channel)
		}

	} else {
		for _, file := range files {
			bytes, err = ioutil.ReadFile(file)
			if debugger.WithError(err, fmt.Sprintf("Unable to read file %s: ", file), err) {
				os.Exit(13)
			}
			handleMessageBytes(bytes, channel)
		}
	}
}

func handleMessageBytes(bytes []byte, channel *amqp.Channel) {
	var err error

	delivery := &DeliveryPlus{}

	err = json.Unmarshal(bytes, &delivery)

	if debugger.WithError(err, "Unable to unmarshal delivery into JSON: ", err) {
		os.Exit(7)
	}

	/*
		DO STUFF WITH INPUT LINE
	*/

	rawDelivery := delivery.RawDelivery
	bodyBytes := rawDelivery.Body

	errorMessage := &ErrorMessage{}

	err = json.Unmarshal(bodyBytes, &errorMessage)
	if debugger.WithError(err, "Unable to unmarshal delivery into JSON: ", err) {
		os.Exit(7)
	}

	oMsg := errorMessage.OriginalMessage

	debugger.Print(fmt.Sprintf("consumed message: %+v", errorMessage.OriginalMessage))

	timestamp, _ := time.Parse(timeFormat, oMsg.Properties.Timestamp)

	msg := &amqp.Publishing{
		ContentType:     oMsg.Properties.ContentType,
		ContentEncoding: oMsg.Properties.ContentEncoding,
		DeliveryMode:    uint8(oMsg.Properties.DeliveryMode),
		CorrelationId:   oMsg.Properties.CorrelationId,
		MessageId:       oMsg.Properties.MessageId,
		Timestamp:       timestamp,
		AppId:           oMsg.Properties.AppId,
		Body:            []byte(oMsg.Payload),
	}

	err = channel.Publish(oMsg.Exchange, "asdf.oMsg", true, false, *msg)
	if debugger.WithError(err, "Unable to publish: ", err) {
		os.Exit(19)
	}
}
