package amqptools

import (
	"time"
)

import (
	"github.com/streadway/amqp"
)

type DeliveryProperties interface {
	GetContentType() string     // MIME content type
	GetContentEncoding() string // MIME content encoding
	GetDeliveryMode() uint8     // queue implemention use - non-persistent (1) or persistent (2)
	GetPriority() uint8         // queue implementation use - 0 to 9
	GetCorrelationId() string   // application use - correlation identifier
	GetReplyTo() string         // application use - address to to reply to (ex: RPC)
	GetExpiration() string      // implementation use - message expiration spec
	GetMessageId() string       // application use - message identifier
	GetTimestamp() time.Time    // application use - message timestamp
	GetType() string            // application use - message type name
	GetUserId() string          // application use - creating user - should be authenticated user
	GetAppId() string           // application use - creating application id
}

func NewAmqpPublishingWithDelivery(dp DeliveryProperties) *amqp.Publishing {
	return &amqp.Publishing{
		Body:            make([]byte, 0),
		ContentType:     dp.GetContentType(),
		ContentEncoding: dp.GetContentEncoding(),
		DeliveryMode:    dp.GetDeliveryMode(),
		Priority:        dp.GetPriority(),
		CorrelationId:   dp.GetCorrelationId(),
		ReplyTo:         dp.GetReplyTo(),
		Expiration:      dp.GetExpiration(),
		MessageId:       dp.GetMessageId(),
		Timestamp:       dp.GetTimestamp(),
		Type:            dp.GetType(),
		UserId:          dp.GetUserId(),
		AppId:           dp.GetAppId(),
	}
}
