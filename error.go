package ifttt

import (
	"errors"

	"github.com/Jeffail/gabs"
)

var (
	// ErrorInvalidToken is the error that you should return if the request was unauthorized
	ErrorInvalidToken = errors.New("Token invalid")
	// ErrorPanicDuringProcess is returned to the server when a panic occurred while handling the request
	ErrorPanicDuringProcess = errors.New("Internal Error")
)

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
