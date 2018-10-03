package ifttt

import (
	"errors"

	"github.com/Jeffail/gabs"
)

var (
	ErrorInvalidToken       = errors.New("Token invalid")
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
