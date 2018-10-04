package ifttt

import (
	"sort"
	"time"

	"github.com/Jeffail/gabs"
)

// TriggerPollRequest represents the request from IFTTT for events regarding this trigger
type TriggerPollRequest struct {
	// TriggerIdentity the identification string of the trigger
	TriggerIdentity string
	// TriggerFields the values of the trigger fields
	TriggerFields map[string]string
	// Limit max number of events requested, you can return a little more but extra events will be ignored
	Limit int
	// User contains the metadata of the IFTTT user (eg: timezone)
	User map[string]string
	// TODO: IFTTT source
}

// TriggerEventCollection a slice of TriggerEvent, the events returned to a trigger poll
type TriggerEventCollection []TriggerEvent

// TriggerEvent represents a single event returned during a trigger poll
// the trigger will be triggered if ALL of the following conditions were met:
// (1) The trigger has NOT been fired by an event with identical IDs
// (2) The event happened later than the time the trigger was created
type TriggerEvent struct {
	// Ingredients contains values of the ingredients of this event
	Ingredients map[string]string
	// Meta contains metadata of this event
	Meta TriggerEventMeta
}

// TriggerEventMeta represents the metadata of the event
type TriggerEventMeta struct {
	// ID should be a unique string identifier to an event, events with different IDs would be considered different events
	ID string
	// Time should be the creation or modification time of the resource that triggers the trigger
	// This field don't have to be constant, but the trigger will not be triggered if the time returned was earlier than the time the trigger was created
	Time time.Time
}

// Trigger is the interface which every registered trigger should implement.
type Trigger interface {
	// Poll askes the trigger for event updates
	// Implementations should return recent events regarding this trigger.
	// It does not matter if you return events which is already reported before, as long as the ID was kept the same.
	Poll(req *TriggerPollRequest, r *Request) (TriggerEventCollection, error)
	// Options is called during a trigger dynamic option request
	// Use req.FieldSlug for the field requested by this request.
	// If your trigger does not contain dynamic options, return nil, nil here.
	Options(req *Request) (*DynamicOption, error)
	// ValidateField is called during a trigger single-field validation request
	// Use req.FieldSlug for the field requested by this request.
	// If the field value is unacceptable, return a non-nil error which contains a friendly message describing the problem
	ValidateField(fieldslug string, value string, req *Request) error
	// ValidateContext is called during a trigger field context validation request
	// You need to contact IFTTT first to have this feature enabled
	// If you do not use this feature, return nil, nil here.
	ValidateContext(values map[string]string, req *Request) (map[string]error, error)
	// RemoveIdentity is called to notify the the trigger identified by triggerid has been deleted and the endpoint can stop tracking updates to this trigger
	// This is for performance reasons, especially when you are using real-time functions.
	// You can safely ignore this by returning nil directly.
	RemoveIdentity(triggerid string) error
	// RealTie should return whether this service supports the IFTTT real-time API.
	// If set to true, IFTTT will poll a lot less often on this trigger and rely on your active notification instead.
	// If you does not know what is IFTTT real-time API, return false here.
	RealTime() bool
}

// Len implements sort.Interface
func (c TriggerEventCollection) Len() int {
	return len(c)
}

// Less implements sort.Interface
func (c TriggerEventCollection) Less(i, j int) bool {
	return c[i].Meta.Time.Unix() > c[j].Meta.Time.Unix()
}

//Swap implements sort.Interface
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
		for key, val := range evt.Ingredients {
			obj.Set(val, key)
		}

		res.ArrayAppend(obj.Data(), "data")
	}
	return res.Bytes()
}
