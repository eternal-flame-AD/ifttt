package ifttt

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/Jeffail/gabs"
	uuid "github.com/satori/go.uuid"
)

// Service described the IFTTT service and handles requests from IFTTT
type Service struct {
	triggers map[string]Trigger
	actions  map[string]Action
	// IFTTT service key used to identify your service
	// get it from you dashboard
	ServiceKey string
	// Healthy should return whether the service is functioning normally
	// Defaults to true
	Healthy func() bool
	// UserInfo should return user info identified by req.UserAccessToken
	// if your service does not require authentication, passing nil should be OK
	UserInfo func(req *Request) (*UserInfo, error)
	logger   *log.Logger
}

func prepareHeader(w http.ResponseWriter) {
	header := w.Header()
	header.Set("Content-Type", "application/json")
}

// RegisterTrigger registers a trigger handler which implements Trigger
func (c *Service) RegisterTrigger(slug string, handler Trigger) {
	if c.triggers == nil {
		c.triggers = make(map[string]Trigger)
	}
	c.triggers[slug] = handler
}

// RegisterAction registers an action handler which implements Action
func (c *Service) RegisterAction(slug string, handler Action) {
	if c.actions == nil {
		c.actions = make(map[string]Action)
	}
	c.actions[slug] = handler
}

// EnableDebug enabled debug output of this service
func (c *Service) EnableDebug() {
	c.logger = log.New(os.Stdout, "IFTTT: ", log.LstdFlags)
}

// ServeHTTP implements http.Handler and handles http requests
func (c Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(500)
			w.Write(marshalError(ErrorPanicDuringProcess, false))
			if c.logger != nil {
				c.logger.Printf("Panic during processing! Stach Trace: %s\n", debug.Stack())
			}
		}
	}()

	if c.actions == nil {
		c.actions = make(map[string]Action)
	}
	if c.triggers == nil {
		c.triggers = make(map[string]Trigger)
	}

	handleError := func(err error) {
		if err == ErrorInvalidToken {
			w.WriteHeader(401)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write(marshalError(err, false))
	}

	prepareHeader(w)

	req, err := parseRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
		return
	}
	req.ServiceRef = &c

	if c.logger != nil {
		c.logger.Printf("Got request %s - %s with type %d\n", r.RequestURI, req.DecodedBody.String(), req.Type)
	}

	// If the request is unauthenticated and the service key is incorrect, refuse to handle it.
	if !req.Authenticated && req.UserAccessToken != c.ServiceKey {
		if c.logger != nil {
			c.logger.Printf("Request %s refused due to incorrect claimed service key: %s", r.RequestURI, req.UserAccessToken)
		}
		handleError(ErrorInvalidToken)
		return
	}

	switch req.Type {
	case ServiceStatus:
		if c.Healthy == nil || c.Healthy() {
			w.WriteHeader(200)
			w.Write([]byte{})
		} else {
			w.WriteHeader(503)
			w.Write([]byte{})
		}
	case UserInfoRequest:
		if c.UserInfo == nil {
			handleError(errors.New("User info not available"))
		} else if info, err := c.UserInfo(req); err != nil {
			handleError(err)
		} else {
			w.WriteHeader(200)
			w.Write(info.marshal())
		}
	case ActionTrigger:
		action, ok := c.actions[req.Slug]
		if !ok {
			handleError(errors.New("Action Not Registered"))
			return
		}
		ahq := &ActionHandleRequest{
			make(map[string]string),
			make(map[string]string),
		}
		if fields, err := req.DecodedBody.S("actionFields").ChildrenMap(); err != nil {
			handleError(err)
			return
		} else {
			for key, val := range fields {
				ahq.ActionFields[key] = val.Data().(string)
			}
		}

		if fields, err := req.DecodedBody.S("user").ChildrenMap(); err != nil {
			handleError(err)
			return
		} else {
			for key, val := range fields {
				ahq.User[key] = val.Data().(string)
			}
		}
		if res, skip, err := action.Handle(ahq, req); err != nil {
			if err == ErrorInvalidToken {
				handleError(err)
			}
			w.WriteHeader(400)
			w.Write(marshalError(err, skip))
			return
		} else {
			w.WriteHeader(200)
			w.Write(res.marshal())
		}
	case TriggerFetch:
		trigger, ok := c.triggers[req.Slug]
		if !ok {
			handleError(errors.New("Trigger Not Registered"))
			return
		}
		tpr := &TriggerPollRequest{
			"",
			make(map[string]string),
			50,
			make(map[string]string),
		}

		tpr.TriggerIdentity = req.DecodedBody.S("trigger_identity").Data().(string)
		if fields, err := req.DecodedBody.S("triggerFields").ChildrenMap(); err != nil {
			handleError(err)
			return
		} else {
			for key, val := range fields {
				tpr.TriggerFields[key] = val.Data().(string)
			}
		}

		if fields, err := req.DecodedBody.S("user").ChildrenMap(); err != nil {
			handleError(err)
			return
		} else {
			for key, val := range fields {
				tpr.User[key] = val.Data().(string)
			}
		}
		if req.DecodedBody.Exists("limit") {
			tpr.Limit = int(req.DecodedBody.S("limit").Data().(float64))
		}
		if evts, err := trigger.Poll(tpr, req); err != nil {
			handleError(err)
			return
		} else {
			if trigger.RealTime() {
				w.Header().Add("X-IFTTT-Realtime", "1")
			}
			w.WriteHeader(200)
			w.Write(evts.marshal())
		}
	case ActionDynamicOptions:
		action, ok := c.actions[req.Slug]
		if !ok {
			handleError(errors.New("Action Not Registered"))
			return
		}
		if options, err := action.Options(req); err != nil {
			handleError(err)
			return
		} else {
			w.WriteHeader(200)
			w.Write(options.marshal())
		}
	case TriggerDynamicOptions:
		trigger, ok := c.triggers[req.Slug]
		if !ok {
			handleError(errors.New("Trigger Not Registered"))
			return
		}

		if options, err := trigger.Options(req); err != nil {
			handleError(err)
			return
		} else {
			w.WriteHeader(200)
			w.Write(options.marshal())
		}
	case TriggerDynamicValidation:
		trigger, ok := c.triggers[req.Slug]
		if !ok {
			handleError(errors.New("Trigger Not Registered"))
			return
		}

		if err := trigger.ValidateField(req.FieldSlug, req.DecodedBody.S("value").Data().(string), req); err != nil {
			w.WriteHeader(200)
			res := gabs.New()
			res.Set(false, "data", "valid")
			res.Set(err.Error(), "data", "message")
			w.Write(res.Bytes())
		} else {
			w.WriteHeader(200)
			res := gabs.New()
			res.Set(true, "data", "valid")
			w.Write(res.Bytes())
		}
	case TriggerContextualValidation:
		trigger, ok := c.triggers[req.Slug]
		if !ok {
			handleError(errors.New("Trigger Not Registered"))
			return
		}
		values := make(map[string]string)
		keymap, err := req.DecodedBody.S("values").ChildrenMap()
		if err != nil {
			handleError(err)
			return
		}
		for key, val := range keymap {
			values[key] = val.Data().(string)
		}

		if ret, err := trigger.ValidateContext(values, req); err != nil {
			handleError(err)
			return
		} else {
			res := gabs.New()
			res.Object("data")
			for key, err := range ret {
				if err == nil {
					res.Set(true, "data", key, "valid")
				} else {
					res.Set(false, "data", key, "valid")
					res.Set(err.Error(), "data", key, "message")
				}
			}
			w.WriteHeader(200)
			w.Write(res.Bytes())
		}
	case TriggerDeleteNotify:
		trigger, ok := c.triggers[req.Slug]
		if !ok {
			handleError(errors.New("Trigger Not Registered"))
			return
		}
		if err := trigger.RemoveIdentity(req.TriggerIdentity); err != nil {
			handleError(err)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte{})
	}

}

// Notify implements the IFTTT realtime API and sends notifications to the IFTTT realtime notification endpoint
func (c *Service) Notify(evt Notification) error {
	req, err := http.NewRequest("POST", "https://realtime.ifttt.com/v1/notifications", bytes.NewReader(evt.marshal()))
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "utf-8")
	req.Header.Set("Content-Type", "application/json")
	uid, err := uuid.NewV1()
	if err != nil {
		return err
	}
	req.Header.Set("X-Request-ID", uid.String())
	req.Header.Set("IFTTT-Service-Key", c.ServiceKey)

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		response, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("Remote returned code %d with: %s", resp.StatusCode, string(response))
	}
	return nil
}
