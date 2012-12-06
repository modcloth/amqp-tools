BIBLIOTECAS := \
  github.com/modcloth/amqp-tools
OBJETIVOS := \
  github.com/modcloth/amqp-tools \
  github.com/modcloth/amqp-tools/amqp-publish-files \
  github.com/modcloth/amqp-tools/amqp-consume

all: build test

build: deps
	go install -x $(OBJETIVOS)

test:
	go test -x -v $(BIBLIOTECAS)

deps:
	go get -x $(OBJETIVOS)

.PHONY: all build deps test
