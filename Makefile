LIBS := amqp-tools
REV_VAR := amqp-tools.RevString
VERSION_VAR := amqp-tools.VersionString
REPO_VERSION := $(shell git describe --always --dirty --tags)
REPO_REV := $(shell git rev-parse --sq HEAD)
GOBUILD_VERSION_ARGS := -ldflags "-X $(REV_VAR) $(REPO_REV) -X $(VERSION_VAR) $(REPO_VERSION)"
JOHNNY_DEPS_VERSION := v0.2.3

all: build test

build: deps
	go install $(GOBUILD_VERSION_ARGS) -x $(LIBS)
	for exe in consume-cat publish-files replay-ninja ; do \
	  go build -o $${GOPATH%%:*}/bin/amqp-$$exe $(GOBUILD_VERSION_ARGS) ./amqp-$$exe ; \
	done

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
