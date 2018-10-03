package ifttt

import "testing"

func TestNotifyMarshal(t *testing.T) {
	not := Notification{}
	not.AddUser("foo")
	not.AddTrigger("baz")

	if res := not.marshal(); !jsonEqual(res, []byte(`{"data":[{"trigger_identity":"baz"},{"user_id":"foo"}]}`)) {
		t.Errorf("MarshalError: Unexpected JSON: %s\n", res)
		t.Fail()
	}
}
