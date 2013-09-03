package amqptools

import "github.com/streadway/amqp"

type deliveryPlus struct {
	RawDelivery amqp.Delivery
	Data        map[string]interface{}
}
