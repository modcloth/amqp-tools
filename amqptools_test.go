package amqptools_test

import (
	"errors"
	"testing"
)

import (
	"github.com/modcloth/amqp-tools"
)

func TestPublishFileResult(t *testing.T) {
	res := &amqptools.PublishFileResult{
		"foo",
		"wat",
		errors.New("not really, nerds!"),
		false,
	}

	if res.Filename != "foo" {
		t.Fail()
	}

	if res.Message != "wat" {
		t.Fail()
	}

	if res.Error.Error() != "not really, nerds!" {
		t.Fail()
	}

	if res.IsFatal {
		t.Fail()
	}
}