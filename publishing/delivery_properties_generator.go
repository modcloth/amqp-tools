package publishing

import (
	"amqp-tools"
	"time"
)

type DeliveryPropertiesGenerator struct {
	ContentType            string
	ContentEncoding        string
	DeliveryMode           uint8
	Priority               uint8
	CorrelationIdGenerator amqptools.Nexter
	ReplyTo                string
	Expiration             string
	MessageIdGenerator     amqptools.Nexter
	Timestamp              int64
	Type                   string
	UserId                 string
	AppId                  string
}
type DeliveryPropertiesHolder struct {
	ContentType            *string
	ContentEncoding        *string
	DeliveryMode           *uint
	Priority               *uint
	CorrelationIdGenerator NexterWrapper
	ReplyTo                *string
	Expiration             *string
	MessageIdGenerator     NexterWrapper
	Timestamp              *int64
	Type                   *string
	UserId                 *string
	AppId                  *string
}

func (me *DeliveryPropertiesHolder) DeliveryPropertiesGenerator() *DeliveryPropertiesGenerator {
	return &DeliveryPropertiesGenerator{
		ContentType:            *me.ContentType,
		ContentEncoding:        *me.ContentEncoding,
		DeliveryMode:           uint8(*me.DeliveryMode),
		Priority:               uint8(*me.Priority),
		CorrelationIdGenerator: &me.CorrelationIdGenerator,
		ReplyTo:                *me.ReplyTo,
		Expiration:             *me.Expiration,
		MessageIdGenerator:     &me.MessageIdGenerator,
		Timestamp:              *me.Timestamp,
		Type:                   *me.Type,
		UserId:                 *me.UserId,
		AppId:                  *me.AppId,
	}
}

func (dph *DeliveryPropertiesGenerator) GetContentType() string     { return dph.ContentType }
func (dph *DeliveryPropertiesGenerator) GetContentEncoding() string { return dph.ContentEncoding }
func (dph *DeliveryPropertiesGenerator) GetDeliveryMode() uint8     { return dph.DeliveryMode }
func (dph *DeliveryPropertiesGenerator) GetPriority() uint8         { return dph.Priority }
func (dph *DeliveryPropertiesGenerator) GetCorrelationId() string {
	result, err := dph.CorrelationIdGenerator.Next()

	if err != nil {
		panic(err)
	}

	return result
}
func (dph *DeliveryPropertiesGenerator) GetReplyTo() string    { return dph.ReplyTo }
func (dph *DeliveryPropertiesGenerator) GetExpiration() string { return dph.Expiration }
func (dph *DeliveryPropertiesGenerator) GetMessageId() string {
	result, err := dph.CorrelationIdGenerator.Next()

	if err != nil {
		panic(err)
	}

	return result
}
func (dph *DeliveryPropertiesGenerator) GetTimestamp() time.Time {
	return time.Unix(dph.Timestamp, 0)
}
func (dph *DeliveryPropertiesGenerator) GetType() string   { return dph.Type }
func (dph *DeliveryPropertiesGenerator) GetUserId() string { return dph.UserId }
func (dph *DeliveryPropertiesGenerator) GetAppId() string  { return dph.AppId }
