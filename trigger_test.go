package ifttt

import (
	"testing"
	"time"
)

func TestTriggerEventCollection(t *testing.T) {
	col := TriggerEventCollection{}

	col = append(col, TriggerEvent{
		Slugs: map[string]string{
			"foo": "bar",
		},
		Meta: TriggerEventMeta{
			ID:   "1",
			Time: time.Unix(100000, 0),
		},
	})

	col = append(col, TriggerEvent{
		Slugs: map[string]string{
			"foo": "bar",
		},
		Meta: TriggerEventMeta{
			ID:   "2",
			Time: time.Unix(200000, 0),
		},
	})

	if res := col.marshal(); !jsonEqual(res, []byte(`{"data":[{"foo":"bar","meta":{"id":"2","timestamp":200000}},{"foo":"bar","meta":{"id":"1","timestamp":100000}}]}`)) {
		t.Errorf("MarshalError: Unexpected JSON: %s\n", res)
		t.Fail()
	}
}
