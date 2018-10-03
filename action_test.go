package ifttt

import "testing"

func TestMarshalActionResult(t *testing.T) {
	res := ActionResult{
		ID: "12345678",
	}
	if res := res.marshal(); !jsonEqual(res, []byte(`{"data":[{"id":"12345678"}]}`)) {
		t.Errorf("MarshalError: Unexpected JSON: %s\n", res)
		t.Fail()
	}
	res = ActionResult{
		ID:  "12345678",
		URL: "https://www.example.com/",
	}
	if res := res.marshal(); !jsonEqual(res, []byte(`{"data":[{"id":"12345678","url":"https://www.example.com/"}]}`)) {
		t.Errorf("MarshalError: Unexpected JSON: %s\n", res)
		t.Fail()
	}
}
