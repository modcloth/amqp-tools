LIBS := \
  github.com/modcloth/amqp-tools
TARGETS := \
  github.com/modcloth/amqp-tools \
  github.com/modcloth/amqp-tools/amqp-publish-files \
  github.com/modcloth/amqp-tools/amqp-consume-cat

all: build test

build: deps
	go install -x $(TARGETS)

test:
	go test -x -v $(LIBS)

deps:
	go get -x $(TARGETS)

.PHONY: all build deps test
