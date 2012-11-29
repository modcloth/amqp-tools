package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"
	"time"
)

import (
	"github.com/jteeuwen/go-pkg-optarg"
	"github.com/streadway/amqp"
)

type ExitCodes int
const (
  Success = 0
  ArgumentParsingError = 1
  ConnectionError = 2
  ChannelOpenError = 3
)


func getOptions() (options *Options) {
	options = NewOptions()

	optarg.UsageInfo = fmt.Sprintf("Usage: %s [options] [exchange] [file ...]:", os.Args[0])

	optarg.Header("Executable options")
	optarg.Add("h", "help", "Help", options.Help)

	optarg.Header("AMQP options")
	optarg.Add("u", "username", "AMQP username", options.Username)
	optarg.Add("p", "password", "AMQP password", options.Password)
	optarg.Add("H", "host", "AMQP host", options.Host)
	optarg.Add("r", "vhost", "AMQP vhost", options.VHost)
	optarg.Add("P", "port", "AMQP port", options.Port)

	optarg.Header("Message options")
	optarg.Add("k", "routingkey", "Routing key", options.RoutingKey)
	optarg.Add("m", "mandatory", "Mandatory", options.Mandatory)
	optarg.Add("i", "immediate", "Immediate", options.Immediate)
	optarg.Add("c", "contenttype", "Content-type", options.ContentType)

	for opt := range optarg.Parse() {
		switch opt.Name {
		case "username":
			options.Username = opt.String()
		case "password":
			options.Password = opt.String()
		case "help":
			options.Help = opt.Bool()
		case "host":
			options.Host = opt.String()
		case "vhost":
			options.VHost = opt.String()
		case "port":
			options.Port = opt.Int()
		case "routingkey":
			options.RoutingKey = opt.String()
		case "mandatory":
			options.Mandatory = opt.Bool()
		case "immediate":
			options.Immediate = opt.Bool()
		case "contenttype":
			options.ContentType = opt.String()
		}
	}

	if options.Help {
		fmt.Println(optarg.UsageString())
		os.Exit(Success)
	}

	if len(optarg.Remainder) < 2 {
		fmt.Printf("Not enough arguments\n%s", optarg.UsageString())
		os.Exit(ArgumentParsingError)
	}

	options.Exchange = optarg.Remainder[0]
	options.Files = optarg.Remainder[1:]

	return
}

type Options struct {
	Help bool

	Username string
	Password string
	Host     string
	VHost    string
	Port     int

	RoutingKey  string
	Mandatory   bool
	Immediate   bool
	ContentType string

	Exchange string
	Files    []string
}

func NewOptions() *Options {
	return &Options{
		Username: "guest",
		Password: "guest",
		Host:     "localhost",
		Port:     5672,
	}
}

func main() {
	var options *Options
	var err error

	var connectionUri string
	var conn *amqp.Connection
	var channel *amqp.Channel
	var message *amqp.Publishing

	options = getOptions()

	connectionUri = fmt.Sprintf("amqp://%s:%s@%s:%d/%s", options.Username, options.Password, options.Host, options.Port, options.VHost)

	if conn, err = amqp.Dial(connectionUri); err != nil {
		fmt.Printf("Failed to open connection: %s", err.Error())
		os.Exit(ConnectionError)
	}
	defer conn.Close()

	if channel, err = conn.Channel(); err != nil {
		log.Fatal("Failed to open channel: ", err)
		os.Exit(ChannelOpenError)
	}

	message = &amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  options.ContentType,
		Body:         make([]byte, 0),
	}

	for _, file := range options.Files {
		if message.ContentType == "" {
			message.ContentType = mime.TypeByExtension(filepath.Ext(file))
		}

		if message.Body, err = ioutil.ReadFile(file); err != nil {
			fmt.Printf("Failed to read file %s: %s", file, err.Error())
			continue
		}

		if err = channel.Publish(options.Exchange, options.RoutingKey, options.Mandatory, options.Immediate, *message); err != nil {
			fmt.Printf("Failed to message file %s: %s", file, err.Error())
		}
	}
}
