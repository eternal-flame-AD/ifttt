package ifttt

import (
	"sort"
	"time"

	"github.com/Jeffail/gabs"
)

type TriggerPollRequest struct {
	TriggerIdentity string
	TriggerFields   map[string]string
	Limit           int
	User            map[string]string
	// TODO: IFTTT source
}

type TriggerEventCollection []TriggerEvent

type TriggerEvent struct {
	Slugs map[string]string
	Meta  TriggerEventMeta
}

type TriggerEventMeta struct {
	ID   string
	Time time.Time
}

type Trigger interface {
	Poll(req *TriggerPollRequest, r *Request) (TriggerEventCollection, error)
	Options(req *Request) (*DynamicOption, error)
	ValidateField(fieldslug string, value string, req *Request) error
	ValidateContext(values map[string]string, req *Request) (map[string]error, error)
	RemoveIdentity(triggerid string) error
	RealTime() bool
}

func (c TriggerEventCollection) Len() int {
	return len(c)
}

func (c TriggerEventCollection) Less(i, j int) bool {
	return c[i].Meta.Time.Unix() > c[j].Meta.Time.Unix()
}

func (c TriggerEventCollection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c TriggerEventCollection) marshal() []byte {
	sort.Sort(c)

	res := gabs.New()
	res.Array("data")
	for _, evt := range c {
		obj := gabs.New()
		obj.Set(evt.Meta.ID, "meta", "id")
		obj.Set(evt.Meta.Time.Unix(), "meta", "timestamp")
		for key, val := range evt.Slugs {
			obj.Set(val, key)
		}

		res.ArrayAppend(obj.Data(), "data")
	}
	return res.Bytes()
}
