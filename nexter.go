package amqptools

import (
	"strconv"
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


func (sp *StaticProvider) Next() (string, error) {
	return sp.Value, nil
}

func (sp *SeriesProvider) Next() (string, error) {
	sp.current++
	return strconv.FormatInt(int64(sp.current), 10), nil
}

}
