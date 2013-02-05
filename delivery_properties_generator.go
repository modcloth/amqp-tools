package amqptools

import (
	"time"
)

type DeliveryPropertiesGenerator struct {
	ContentType            string
	ContentEncoding        string
	DeliveryMode           uint
	Priority               uint
	CorrelationIdGenerator Nexter
	ReplyTo                string
	Expiration             string
	MessageIdGenerator     Nexter
	Timestamp              int64
	Type                   string
	UserId                 string
	AppId                  string
}

func (dph *DeliveryPropertiesGenerator) GetContentType() string     { return dph.ContentType }
func (dph *DeliveryPropertiesGenerator) GetContentEncoding() string { return dph.ContentEncoding }
func (dph *DeliveryPropertiesGenerator) GetDeliveryMode() uint8     { return uint8(dph.DeliveryMode) }
func (dph *DeliveryPropertiesGenerator) GetPriority() uint8         { return uint8(dph.Priority) }
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
