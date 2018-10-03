package ifttt

import (
	"encoding/json"
	"reflect"
	"testing"
)

// https://gist.github.com/turtlemonvh/e4f7404e28387fadb8ad275a99596f67
func jsonEqual(a, b []byte) bool {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal(a, &o1)
	if err != nil {
		return false
	}
	err = json.Unmarshal(b, &o2)
	if err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}

func TestDynamicOption(t *testing.T) {
	opt := new(DynamicOption)
	opt.AddString("foo", "123")

	bar := new(DynamicOption)
	bar.AddString("baz", "456")
	opt.AddCategory("bar", bar)

	if res := opt.marshal(); !jsonEqual(res, []byte(`{"data":[{"label":"bar","values":[{"label":"baz","value":"456"}]},{"label":"foo","value":"123"}]}`)) && !jsonEqual(res, []byte(`{"data":[{"label":"foo","value":"123"},{"label":"bar","values":[{"label":"baz","value":"456"}]}]}`)) {
		t.Errorf("MarshalError: Unexpected JSON: %s\n", res)
		t.Fail()
	}
}
