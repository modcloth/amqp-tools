package main

import (
	"flag"
)

import (
	. "github.com/modcloth/amqp-tools"
)

var (
	uri = flag.String("U", "amqp://guest:guest@localhost:5672", "AMQP Connection URI")
)

func main() {
	flag.Parse()

	TailRabbitLogs(*uri)
}
