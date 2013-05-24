package amqptools

import (
	"fmt"
	"os"
	"path"
)

var (
	VersionString string
	progName      string
)

func init() {
	progName = path.Base(os.Args[0])
}

func PrintVersion() {
	if VersionString == "" {
		VersionString = "<unknown>"
	}
	fmt.Printf("%s %s\n", progName, VersionString)
	return
}
