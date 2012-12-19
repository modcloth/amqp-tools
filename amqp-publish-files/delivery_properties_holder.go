package main

import (
	"time"
)

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

func (dph *DeliveryPropertiesHolder) GetContentType() string     { return *dph.ContentType }
func (dph *DeliveryPropertiesHolder) GetContentEncoding() string { return *dph.ContentEncoding }
func (dph *DeliveryPropertiesHolder) GetDeliveryMode() uint8     { return uint8(*dph.DeliveryMode) }
func (dph *DeliveryPropertiesHolder) GetPriority() uint8         { return uint8(*dph.Priority) }
func (dph *DeliveryPropertiesHolder) GetCorrelationId() string {
  result, err := dph.CorrelationIdGenerator.Next()

  if err != nil {
	panic(err)
  }

  return result
}
func (dph *DeliveryPropertiesHolder) GetReplyTo() string    { return *dph.ReplyTo }
func (dph *DeliveryPropertiesHolder) GetExpiration() string { return *dph.Expiration }
func (dph *DeliveryPropertiesHolder) GetMessageId() string {
  result, err := dph.CorrelationIdGenerator.Next()

  if err != nil {
	panic(err)
  }

  return result
}
func (dph *DeliveryPropertiesHolder) GetTimestamp() time.Time {
	return time.Unix(*dph.Timestamp, 0)
}
func (dph *DeliveryPropertiesHolder) GetType() string   { return *dph.Type }
func (dph *DeliveryPropertiesHolder) GetUserId() string { return *dph.UserId }
func (dph *DeliveryPropertiesHolder) GetAppId() string  { return *dph.AppId }
