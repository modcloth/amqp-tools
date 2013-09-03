package main

import (
	"flag"
	"fmt"
	"os"
	"path"
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

var (
	uriFlag     = flag.String("U", "amqp://guest:guest@localhost:5672", "AMQP Connection URI")
	rmqLogsFlag = flag.Bool("rabbitmq.logs", false, "Consume from amq.rabbitmq.logs and amq.rabbitmq.trace")
	showCatFlag = flag.Bool("mrow", false, "")
	versionFlag = flag.Bool("version", false, "Print version and exit")
	revFlag     = flag.Bool("rev", false, "Print git revision and exit")
	quit        = make(chan bool)

	queueBindings QueueBindings
	debugger      Debugger
)

func init() {
	flag.Var(&queueBindings, "q", "Queue bindings specified as \"/\"-delimited strings of the form \"exchange/queue-name/routing-key\"")
	flag.Var(&debugger, "debug", "Show debug output")
}

func main() {
	flag.Parse()

	if *showCatFlag {
		fmt.Println(CONSUME_CAT)
		os.Exit(0)
	}

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

	deliveries := make(chan interface{})

	if len(queueBindings) > 0 {
		for _, binding := range queueBindings {
			debugger.Print(fmt.Sprintf("Binding to %s", binding))
		}

		go ConsumeForBindings(*uriFlag, queueBindings, deliveries, debugger)

		go func() {
			for delivery := range deliveries {
				switch delivery.(type) {
				case nil:
					debugger.Print("Done consuming. Thanks for playing!")
					quit <- true
				default:
					HandleDelivery(delivery.(amqp.Delivery), debugger)
				}
			}
		}()
	} else {
		fmt.Println("ERROR: define at least one exchange/queue/binding argument.")
		flag.Usage()
		os.Exit(NOT_COOL_ZEUS)
	}
	<-quit
}
