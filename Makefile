LIBS := amqp-tools
REV_VAR := amqp-tools.RevString
VERSION_VAR := amqp-tools.VersionString
REPO_VERSION := $(shell git describe --always --dirty --tags)
REPO_REV := $(shell git rev-parse --sq HEAD)
GOBUILD_VERSION_ARGS := -ldflags "-X $(REV_VAR) $(REPO_REV) -X $(VERSION_VAR) $(REPO_VERSION)"
JOHNNY_DEPS_VERSION := v0.1.3

all: build test

build: deps
	go install $(GOBUILD_VERSION_ARGS) -x $(LIBS)
	go build -o $${GOPATH%%:*}/bin/amqp-consume-cat $(GOBUILD_VERSION_ARGS) ./amqp-consume-cat
	go build -o $${GOPATH%%:*}/bin/amqp-publish-files $(GOBUILD_VERSION_ARGS) ./amqp-publish-files
	go build -o $${GOPATH%%:*}/bin/amqp-replay-ninja $(GOBUILD_VERSION_ARGS) ./amqp-replay-ninja

test:
	go test $(GOBUILD_VERSION_ARGS) -x -v $(LIBS)

deps: johnny_deps
	if [ ! -L $${GOPATH%%:*}/src/amqp-tools ] ; then gvm linkthis ; fi
	./johnny_deps

johnny_deps:
	curl -s -o $@ https://raw.github.com/VividCortex/johnny-deps/$(JOHNNY_DEPS_VERSION)/bin/johnny_deps
	chmod +x $@

clean:
	go clean -X $(LIBS) || true
	if [ -d $${GOPATH%%:*}/pkg ] ; then \
	  find $${GOPATH%%:*}/pkg -name '*amqp-tools*' -exec rm -v {} \; ; \
	fi

distclean: clean
	rm -f ./johnny_deps

.PHONY: all build test deps johnny_deps clean distclean
