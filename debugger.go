package amqptools

import (
	"log"
	"os"
)

type Debugger struct {
	debugOn bool
}

func (d *Debugger) DebugOn() {
	d.debugOn = true
}

func (d *Debugger) DebugOff() {
	d.debugOn = false
}

func (d *Debugger) SetDebugOn(val bool) {
	d.debugOn = val
}

func (d *Debugger) Print(message string) {
	if d.debugOn {
		log.Println(message)
	}
}

func (d *Debugger) WithError(err error, message string) bool {
	if err != nil {
		d.Print(message)
		return true
	}
	return false
}

func (d *Debugger) Fatal(err error, message string) {
	if err != nil {
		d.Print(message)
		os.Exit(1)
	}
}
