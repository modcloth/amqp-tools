package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

import (
	"github.com/jteeuwen/go-pkg-optarg"
	. "github.com/modcloth/amqp-tools"
)

type ExitCodes int

const (
	Success              = 0
	ArgumentParsingError = 1
	FatalError           = 2
)

func getOptions() (options *Options) {
	options = NewOptions()

	optarg.UsageInfo = fmt.Sprintf(
		"Usage: %s [options] [exchange] [file ...]:", os.Args[0])

	optarg.Header("Executable options")
	optarg.Add("h", "help", "Help", options.Help)

	optarg.Header("AMQP options")
	optarg.Add("U", "uri", "AMQP connection URI", options.URI)
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
		case "uri":
			options.URI = opt.String()
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

	URI      string
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
		URI:      "",
		Username: "guest",
		Password: "guest",
		Host:     "localhost",
		Port:     5672,
	}
}

func main() {
	var options *Options

	options = getOptions()

	connectionUri := options.URI
	if len(connectionUri) < 1 {
		connectionUri = fmt.Sprintf("amqp://%s:%s@%s:%d/%s", options.Username,
			options.Password, options.Host, options.Port, options.VHost)
	}

	fileChan := make(chan string)
	resultChan := make(chan *PublishFileResult)

	go func() {
		defer close(fileChan)

		if len(options.Files) == 1 && options.Files[0] == "-" {
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
			for _, file := range options.Files {
				fileChan <- file
			}
		}
	}()

	go PublishFiles(fileChan, connectionUri, options.ContentType,
		options.Exchange, options.RoutingKey, options.Mandatory,
		options.Immediate, resultChan)

	for result := range resultChan {
		if result.Error != nil {
			if result.IsFatal {
				log.Println("FATAL:", result.Message, result.Error)
				os.Exit(FatalError)
			} else {
				log.Println("ERROR:", result.Message,
					result.Filename, result.Error)
			}
		} else {
			log.Println(result.Message, result.Filename)
		}
	}
}
