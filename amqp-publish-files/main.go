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

var (
	amqpUri      = flag.String("U", "", "AMQP connection URI")
	amqpUsername = flag.String("u", "guest", "AMQP username")
	amqpPassword = flag.String("p", "guest", "AMQP password")
	amqpHost     = flag.String("H", "localhost", "AMQP host")
	amqpVHost    = flag.String("r", "", "AMQP vhost")
	amqpPort     = flag.Int("P", 5672, "AMQP port")

	routingKey = flag.String("k", "", "Publish message to routing key")
	mandatory  = flag.Bool("m", false,
		"Publish message with mandatory property set.")
	immediate = flag.Bool("i", false,
		"Publish message with immediate property set.")
	contentType = flag.String("c", "",
		"Default content type, else derived from file extension.")

	usageString = `Usage: %s [options] <exchange> <file> [file file ...]

Publishes files as messages to a given exchange.  If there is only a single
filename entry and it is "-", then file names will be read from standard
input assuming entries delimited by at least a line feed ("\n").  Any extra
whitespace in each entry will be stripped before attempting to open the file.

`
)

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
	resultChan := make(chan *PublishFileResult)

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

	go PublishFiles(fileChan, connectionUri, *contentType, exchange,
		*routingKey, *mandatory, *immediate, resultChan)

	for result := range resultChan {
		if result.Error != nil {
			if result.IsFatal {
				log.Println("FATAL:", result.Message, result.Error)
				os.Exit(FATAL_ERROR)
			} else {
				log.Println("ERROR:", result.Message,
					result.Filename, result.Error)
				hadError = true
			}
		} else {
			log.Println(result.Message, result.Filename)
		}
	}

	if hadError {
		os.Exit(PARTIAL_FAILURE)
	} else {
		os.Exit(SUCCESS)
	}
}
