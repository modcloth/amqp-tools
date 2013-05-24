LIBS := \
  github.com/modcloth/amqp-tools
TARGETS := \
  github.com/modcloth/amqp-tools \
  github.com/modcloth/amqp-tools/amqp-publish-files \
  github.com/modcloth/amqp-tools/amqp-consume-cat
VERSION_VAR := github.com/modcloth/amqp-tools.VersionString
REPO_VERSION := $(shell git describe --always --dirty --tags)
GOBUILD_VERSION_ARGS := -ldflags "-X $(VERSION_VAR) $(REPO_VERSION)"

all: build test

build: deps
	go install $(GOBUILD_VERSION_ARGS) -x $(TARGETS)

test:
	go test $(GOBUILD_VERSION_ARGS) -x -v $(LIBS)

deps:
	go get $(GOBUILD_VERSION_ARGS) -x $(TARGETS)

.PHONY: all build deps test
