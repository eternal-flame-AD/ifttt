package ifttt

import "github.com/Jeffail/gabs"

type ActionResult struct {
	ID  string
	URL string
}

type ActionHandleRequest struct {
	ActionFields map[string]string
	User         map[string]string
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

type Action interface {
	Options(req *Request) (*DynamicOption, error)
	Handle(r *ActionHandleRequest, req *Request) (res *ActionResult, skip bool, err error)
}
