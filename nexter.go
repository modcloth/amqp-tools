package amqptools

import "strconv"

type Nexter interface {
	Next() string
}

type SeriesProvider struct {
	current int
}

type StaticProvider struct {
	Value string
}

func (sp *StaticProvider) Next() string {
	return sp.Value
}

func (sp *SeriesProvider) Next() string {
	sp.current++
	return strconv.FormatInt(int64(sp.current), 10)
}
