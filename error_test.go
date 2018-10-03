package ifttt

import (
	"errors"
	"testing"
)

func TestMarshalError(t *testing.T) {
	err := errors.New("test error")

	if res := marshalError(err, false); !jsonEqual(res, []byte(`{"errors":[{"message":"test error"}]}`)) {
		t.Errorf("MarshalError: Unexpected JSON: %s\n", res)
		t.Fail()
	}

	if res := marshalError(err, true); !jsonEqual(res, []byte(`{"errors":[{"message":"test error","status":"SKIP"}]}`)) {
		t.Errorf("MarshalError: Unexpected JSON: %s\n", res)
		t.Fail()
	}
}
