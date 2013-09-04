package publishing

import "github.com/modcloth/amqp-tools"

type NexterWrapper struct{ nexter amqptools.Nexter }

func (nw *NexterWrapper) Next() (string, error) {
	if nw.nexter == nil {
		nw.nexter = new(amqptools.UUIDProvider)
	}
	return nw.nexter.Next()
}
func (nw *NexterWrapper) String() string { return "uuid" }
func (nw *NexterWrapper) Set(arg string) error {
	switch arg {
	case "uuid":
		nw.nexter = new(amqptools.UUIDProvider)
	case "series":
		nw.nexter = new(amqptools.SeriesProvider)
	default:
		nw.nexter = &amqptools.StaticProvider{
			Value: arg,
		}
	}
	return nil
}
