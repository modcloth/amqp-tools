package amqptools

type NexterWrapper struct{ nexter Nexter }

func (nw *NexterWrapper) Next() (string, error) {
	if nw.nexter == nil {
		nw.nexter = new(UUIDProvider)
	}
	return nw.nexter.Next()
}
func (nw *NexterWrapper) String() string { return "uuid" }
func (nw *NexterWrapper) Set(arg string) error {
	switch arg {
	case "uuid":
		nw.nexter = new(UUIDProvider)
	case "series":
		nw.nexter = new(SeriesProvider)
	default:
		nw.nexter = &StaticProvider{
			Value: arg,
		}
	}
	return nil
}
