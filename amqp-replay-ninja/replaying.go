package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

import (
	"amqp-tools"
	"github.com/streadway/amqp"
)

const timeFormat = "Mon Jan 02 15:04:05 MST 2006"

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

func HandleMessageBytes(bytes []byte, channel *amqp.Channel, debugger amqptools.Debugger) {
	var err error

	delivery := &amqptools.DeliveryPlus{}

	err = json.Unmarshal(bytes, &delivery)

	if debugger.WithError(err, "Unable to unmarshal delivery into JSON: ", err) {
		os.Exit(7)
	}

	// DO STUFF WITH INPUT LINE

	rawDelivery := delivery.RawDelivery
	bodyBytes := rawDelivery.Body

	errorMessage := &ErrorMessage{}

	err = json.Unmarshal(bodyBytes, &errorMessage)
	if debugger.WithError(err, "Unable to unmarshal delivery into JSON: ", err) {
		os.Exit(7)
	}

	oMsg := errorMessage.OriginalMessage

	debugger.Print(fmt.Sprintf("consumed message: %+v", errorMessage.OriginalMessage))

	var timestamp time.Time

	if *timeNow {
		timestamp = time.Now().UTC()
	} else {
		timestamp, err = time.Parse(timeFormat, oMsg.Properties.Timestamp)
		if debugger.WithError(err, fmt.Sprintf("Unable to parse timestamp: %+v", oMsg.Properties.Timestamp), err) {
			os.Exit(11)
		}
	}

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

	err = channel.Publish(oMsg.Exchange, oMsg.RoutingKey, true, false, *msg)
	if debugger.WithError(err, "Unable to publish: ", err) {
		os.Exit(19)
	}
}
