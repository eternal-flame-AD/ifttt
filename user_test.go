package ifttt

import "testing"

func TestUserMarshal(t *testing.T) {
	user := UserInfo{
		Name: "foo",
		ID:   "123",
	}

	if res := user.marshal(); !jsonEqual(res, []byte(`{"data":{"id":"123","name":"foo"}}`)) {
		t.Errorf("MarshalError: Unexpected JSON: %s\n", res)
		t.Fail()
	}

	user.URL = "https://www.example.com/"

	if res := user.marshal(); !jsonEqual(res, []byte(`{"data":{"id":"123","name":"foo","url":"https://www.example.com/"}}`)) {
		t.Errorf("MarshalError: Unexpected JSON: %s\n", res)
		t.Fail()
	}
}
