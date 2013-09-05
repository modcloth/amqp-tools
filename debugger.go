package amqptools

import (
	"log"
	"os"
)

type Debugger struct {
	on bool
}

func (d *Debugger) On() {
	d.on = true
}

func (d *Debugger) Off() {
	d.on = false
}

func (d *Debugger) SetDebug(val bool) {
	d.on = val
}

func (d *Debugger) String() string {
	return ""
}

func (d *Debugger) Set(value string) error {
	if value == "true" {
		d.On()
	} else {
		d.Off()
	}
	return nil
}

func (d *Debugger) IsBoolFlag() bool {
	return true
}

func (d *Debugger) Print(message ...interface{}) {
	if d.on {
		log.Println(message...)
	}
}

func (d *Debugger) WithError(err error, message ...interface{}) bool {
	if err != nil {
		d.Print(message...)
		return true
	}
	return false
}

func (d *Debugger) Fatal(err error, message ...interface{}) {
	if err != nil {
		d.Print(message...)
		os.Exit(1)
	}
}
