package ifttt

import (
	"errors"

	"github.com/Jeffail/gabs"
)

var (
	// ErrorPanicDuringProcess is returned to the server when a panic occurred while handling the request
	ErrorPanicDuringProcess = errors.New("Internal Error")
)

// AuthError is returned as an error when the request failed to satisfy authentication requirements
// An optional Message can be defined to return to the user as an explanation of the error, defaults to "Token invalid"
type AuthError struct {
	Message string
}

func (c AuthError) Error() string {
	if len(c.Message) > 0 {
		return c.Message
	}
	return "Token invalid"
}

func marshalError(err error, skip bool) []byte {
	data := gabs.New()
	errObj := gabs.New()
	errObj.Set(err.Error(), "message")
	if skip {
		errObj.Set("SKIP", "status")
	}
	data.Array("errors")
	data.ArrayAppend(errObj.Data(), "errors")
	return data.Bytes()
}
