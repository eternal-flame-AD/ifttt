package ifttt

import "github.com/Jeffail/gabs"

// ActionResult returns the result of an activity
// https://platform.ifttt.com/docs/api_reference#actions
type ActionResult struct {
	// ID a string value which uniquely identifies the resource created or modified by the action.
	ID string
	// URL optional parameter, URL to the resource created or modified by the action.
	URL string
}

// ActionHandleRequest describes a request to handle an action
// https://platform.ifttt.com/docs/api_reference#actions
type ActionHandleRequest struct {
	// ActionFields are the parameters set in an applet
	ActionFields map[string]string
	// User contains the metadata of the IFTTT user (eg: timezone)
	User map[string]string
	// TODO: IFTTT source
}

func (c *ActionResult) marshal() []byte {
	res := gabs.New()
	res.Array("data")
	obj := gabs.New()
	obj.Set(c.ID, "id")
	if len(c.URL) > 0 {
		obj.Set(c.URL, "url")
	}
	res.ArrayAppend(obj.Data(), "data")
	return res.Bytes()
}

// Action is the interface which every registered action should implement.
type Action interface {
	// Options is called during an action dynamic option request
	// Use req.FieldSlug for the field requested by this action.
	// If your action does not contain dynamic options, return nil, nil here.
	Options(req *Request) (*DynamicOption, error)
	// Handle is called then an applet triggered this action.
	// This is where you should check the validity of the request and handle the action.
	// If the request was unauthorized, return ifttt.ErrorInvalidToken
	// If the request cannot be performed due to a temporary error and you want IFTTT to retry this action later, set skip to false and return the error.
	// If there is a problem with the request that you cannot handle(eg: conflicting parameters), set skip to true and return the error, IFTTT will notify the user with the description of your error and give up.
	Handle(r *ActionHandleRequest, req *Request) (res *ActionResult, skip bool, err error)
}
