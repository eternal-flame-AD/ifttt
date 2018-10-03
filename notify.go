package ifttt

import (
	"errors"

	"github.com/Jeffail/gabs"
)

type Notification struct {
	users    []string
	triggers []string
}

func (c *Notification) len() int {
	s := 0
	if c.users != nil {
		s += len(c.users)
	}
	if c.triggers != nil {
		s += len(c.triggers)
	}
	return s
}
func (c *Notification) AddUser(userid string) error {
	if c.users == nil {
		c.users = make([]string, 0)
	}
	if c.len() >= 100 {
		return errors.New("Too many notifications. Try splitting them")
	}
	c.users = append(c.users, userid)
	return nil
}

func (c *Notification) AddTrigger(triggerIdent string) error {
	if c.triggers == nil {
		c.triggers = make([]string, 0)
	}
	if c.len() >= 100 {
		return errors.New("Too many notifications. Try splitting them")
	}
	c.triggers = append(c.triggers, triggerIdent)
	return nil
}

func (c *Notification) marshal() []byte {
	res := gabs.New()
	res.Array("data")

	if c.triggers != nil {
		for _, val := range c.triggers {
			obj := gabs.New()
			obj.Set(val, "trigger_identity")
			res.ArrayAppend(obj.Data(), "data")
		}
	}

	if c.users != nil {
		for _, val := range c.users {
			obj := gabs.New()
			obj.Set(val, "user_id")
			res.ArrayAppend(obj.Data(), "data")
		}
	}

	return res.Bytes()
}
