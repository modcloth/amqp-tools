TARGETS := \
  github.com/modcloth/amqp-tools/amqp-publish-file

all: build test

build: deps
	go install -x $(TARGETS)

test:
	go test -x -v $(TARGETS)

deps:
	go get -x $(TARGETS)

.PHONY: all build deps test
