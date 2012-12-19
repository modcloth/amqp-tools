package amqptools

import (
	"strconv"
)

import (
	uuid "github.com/nu7hatch/gouuid"
)

type Nexter interface {
	Next() (string, error)
}

type SeriesProvider struct {
	current int
}

type StaticProvider struct {
	Value string
}

type UUIDProvider struct{}

func (sp *StaticProvider) Next() (string, error) {
	return sp.Value, nil
}

func (sp *SeriesProvider) Next() (string, error) {
	sp.current++
	return strconv.FormatInt(int64(sp.current), 10), nil
}

func (up *UUIDProvider) Next() (string, error) {
  uuid, err := uuid.NewV4()

  if err != nil {
	return "", err
  }

  return uuid.String(), nil
}
