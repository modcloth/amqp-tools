package amqptools_test

import (
	"errors"
	"testing"
)

import (
	"amqp-tools/publishing"
)

func TestPublishFileResult(t *testing.T) {
	res := &publishing.PublishResult{"wat", errors.New("not really, nerds!"), false}

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
