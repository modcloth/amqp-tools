package amqptools

import "github.com/streadway/amqp"

type DeliveryPlus struct {
	RawDelivery amqp.Delivery
	Data        map[string]interface{}
}

type PublishResult struct {
	Message string
	Error   error
	IsFatal bool
}
