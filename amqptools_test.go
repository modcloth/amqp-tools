package amqptools_test

import (
	"amqp-tools"
	"errors"
	"testing"
)

func TestPublishFileResult(t *testing.T) {
	res := &amqptools.PublishResult{"wat", errors.New("not really, nerds!"), false}

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
