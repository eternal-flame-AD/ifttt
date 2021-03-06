package ifttt

import "github.com/Jeffail/gabs"

// UserInfo represents the user info returned to the user info request
type UserInfo struct {
	Name string
	ID   string
	URL  string
}

func (c *UserInfo) marshal() []byte {
	res := gabs.New()
	res.Set(c.Name, "data", "name")
	res.Set(c.ID, "data", "id")
	if len(c.URL) > 0 {
		res.Set(c.URL, "data", "url")
	}
	return res.Bytes()
}
