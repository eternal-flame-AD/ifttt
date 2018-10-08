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

func TestAuthError(t *testing.T) {
	isAuthError := func(err error) bool {
		_, ok := err.(AuthError)
		return ok
	}

	if !isAuthError(AuthError{"test"}) {
		t.Errorf("Failed to parse auth error")
		t.Fail()
	}

	if !isAuthError(AuthError{}) {
		t.Errorf("Failed to parse auth error")
		t.Fail()
	}

	if isAuthError(errors.New("test")) {
		t.Errorf("Failed to parse auth error")
		t.Fail()
	}

}
