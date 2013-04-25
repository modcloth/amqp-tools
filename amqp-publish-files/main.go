package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

import (
	. "github.com/modcloth/amqp-tools"
)

const (
	SUCCESS           = 0
	ARG_PARSING_ERROR = 1
	FATAL_ERROR       = 86
	PARTIAL_FAILURE   = 9
)

type DeliveryPropertiesHolder struct {
	ContentType            *string
	ContentEncoding        *string
	DeliveryMode           *uint
	Priority               *uint
	CorrelationIdGenerator NexterWrapper
	ReplyTo                *string
	Expiration             *string
	MessageIdGenerator     NexterWrapper
	Timestamp              *int64
	Type                   *string
	UserId                 *string
	AppId                  *string
}

func (me *DeliveryPropertiesHolder) DeliveryPropertiesGenerator() *DeliveryPropertiesGenerator {
	return &DeliveryPropertiesGenerator{
		ContentType:            *me.ContentType,
		ContentEncoding:        *me.ContentEncoding,
		DeliveryMode:           uint8(*me.DeliveryMode),
		Priority:               uint8(*me.Priority),
		CorrelationIdGenerator: &me.CorrelationIdGenerator,
		ReplyTo:                *me.ReplyTo,
		Expiration:             *me.Expiration,
		MessageIdGenerator:     &me.MessageIdGenerator,
		Timestamp:              *me.Timestamp,
		Type:                   *me.Type,
		UserId:                 *me.UserId,
		AppId:                  *me.AppId,
	}
}

var (
	deliveryProperties *DeliveryPropertiesHolder = new(DeliveryPropertiesHolder)
	amqpUri                                      = flag.String("uri", "", "AMQP connection URI")
	amqpUsername                                 = flag.String("user", "guest", "AMQP username")
	amqpPassword                                 = flag.String("password", "guest", "AMQP password")
	amqpHost                                     = flag.String("host", "localhost", "AMQP host")
	amqpVHost                                    = flag.String("vhost", "", "AMQP vhost")
	amqpPort                                     = flag.Int("port", 5672, "AMQP port")

	routingKey = flag.String("routing-key", "", "Publish message to routing key")
	mandatory  = flag.Bool("mandatory", false,
		"Publish message with mandatory property set.")
	immediate = flag.Bool("immediate", false,
		"Publish message with immediate property set.")
	numRoutines = flag.Int("threads", 3, "Number of concurrent publishers")

	usageString = `Usage: %s [options] <exchange> <file> [file file ...]

Publishes files as messages to a given exchange.  If there is only a single
filename entry and it is "-", then file names will be read from standard
input assuming entries delimited by at least a line feed ("\n").  Any extra
whitespace in each entry will be stripped before attempting to open the file.

`
)

func init() {
	deliveryProperties.ContentType = flag.String("content-type", "",
		"Content-type, else derived from file extension.")
	deliveryProperties.ContentEncoding = flag.String("content-encoding", "UTF-8",
		"Mime content-encoding.")
	deliveryProperties.DeliveryMode = flag.Uint("delivery-mode", 2,
		"Delivery mode (1 for non-persistent, 2 for persistent.")
	deliveryProperties.Priority = flag.Uint("priority", 0, "queue implementation use - 0 to 9")
	deliveryProperties.ReplyTo = flag.String("replyto", "", "application use - address to to reply to (ex: rpc)")
	deliveryProperties.Expiration = flag.String("expiration", "", "implementation use - message expiration spec")
	deliveryProperties.Timestamp = flag.Int64("timestamp", time.Now().Unix(), "unix timestamp of message")
	deliveryProperties.Type = flag.String("type", "", "application use - message type name")
	deliveryProperties.UserId = flag.String("userid", "", "application use - creating user - should be authenticated user")
	deliveryProperties.AppId = flag.String("appid", "", "application use - creating application id")

	flag.Var(&deliveryProperties.CorrelationIdGenerator,
		"correlationid",
		"'series' for incrementing ids, 'uuid' for UUIDs, static value otherwise")
	flag.Var(&deliveryProperties.MessageIdGenerator,
		"messageid",
		"'series' for incrementing ids, 'uuid' for UUIDs, static value otherwise")
}

type NexterWrapper struct{ nexter Nexter }

func (nw *NexterWrapper) Next() (string, error) {
	if nw.nexter == nil {
		nw.nexter = new(UUIDProvider)
	}
	return nw.nexter.Next()
}
func (nw *NexterWrapper) String() string { return "uuid" }
func (nw *NexterWrapper) Set(arg string) error {
	switch arg {
	case "uuid":
		nw.nexter = new(UUIDProvider)
	case "series":
		nw.nexter = new(SeriesProvider)
	default:
		nw.nexter = &StaticProvider{
			Value: arg,
		}
	}
	return nil
}

func main() {
	hadError := false

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageString, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Fprintf(os.Stderr,
			"ERROR: The exchange name and a list of file names are required\n")
		flag.Usage()
		os.Exit(ARG_PARSING_ERROR)
	}

	exchange := flag.Arg(0)
	files := flag.Args()[1:flag.NArg()]

	connectionUri := *amqpUri
	if len(connectionUri) < 1 {
		connectionUri = fmt.Sprintf("amqp://%s:%s@%s:%d/%s", *amqpUsername,
			*amqpPassword, *amqpHost, *amqpPort, *amqpVHost)
	}

	fileChan := make(chan string)
	resultChan := make(chan *PublishResult)

	go func() {
		defer close(fileChan)

		if len(files) == 1 && files[0] == "-" {
			log.Println("Reading files from stdin")
			stdin := bufio.NewReader(os.Stdin)
			for {
				line, err := stdin.ReadString('\n')
				if err != nil {
					if err != io.EOF {
						log.Println("ERROR:", err)
					}
					break
				}
				fileChan <- strings.TrimSpace(line)
			}
		} else {
			log.Println("Using files provided on command line")
			for _, file := range files {
				fileChan <- file
			}
		}
	}()

	for i := 0; i < *numRoutines; i++ {
		go PublishFiles(fileChan, connectionUri, exchange, *routingKey,
			*mandatory, *immediate, deliveryProperties.DeliveryPropertiesGenerator(), resultChan)
	}

	for result := range resultChan {
		if result.Error != nil {
			if result.IsFatal {
				log.Println("FATAL:", result.Message, result.Error)
				os.Exit(FATAL_ERROR)
			} else {
				log.Println("ERROR:", result.Message, result.Error)
				hadError = true
			}
		} else {
			log.Println(result.Message)
		}
	}

	if hadError {
		os.Exit(PARTIAL_FAILURE)
	} else {
		os.Exit(SUCCESS)
	}
}
