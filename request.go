package ifttt

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/Jeffail/gabs"
)

var (
	urlRegexpTrigger                      = regexp.MustCompile("^/ifttt/v1/triggers/([^/]+)$")
	urlRegexpTriggerIdentity              = regexp.MustCompile("^/ifttt/v1/triggers/([^/]+)/trigger_identity/([^/]+)$")
	urlRegexpTriggerFieldOption           = regexp.MustCompile("^/ifttt/v1/triggers/([^/]+)/fields/([^/]+)/options$")
	urlRegexpTriggerFieldOptionValidation = regexp.MustCompile("^/ifttt/v1/triggers/([^/]+)/fields/([^/]+)/validate$")
	urlRegexpTriggerContextualValidation  = regexp.MustCompile("^/ifttt/v1/triggers/([^/]+)/validate$")
	urlRegexpAction                       = regexp.MustCompile("^/ifttt/v1/actions/([^/]+)$")
	urlRegexpActionFieldOption            = regexp.MustCompile("^/ifttt/v1/actions/([^/]+)/fields/([^/]+)/options$")
)

type RequestType int

const (
	Unknown RequestType = iota
	TriggerFetch
	TriggerDeleteNotify
	TriggerDynamicOptions
	TriggerDynamicValidation
	TriggerContextualValidation
	ActionTrigger
	ActionDynamicOptions
	ServiceStatus
	UserInfoRequest
)

type Request struct {
	Authenticated   bool
	UserAccessToken string
	RequestUUID     string
	Slug            string
	FieldSlug       string
	TriggerIdentity string
	DecodedBody     *gabs.Container
	Type            RequestType
	RawRequest      *http.Request
}

func parseRequest(r *http.Request) (*Request, error) {

	res := &Request{
		RequestUUID: r.Header.Get("X-Request-ID"),
		RawRequest:  r,
	}

	authheader := r.Header.Get("Authorization")
	if authheader != "" {
		res.Authenticated = true
		res.UserAccessToken = strings.TrimPrefix(authheader, "Bearer ")
	} else {
		res.Authenticated = false
		res.UserAccessToken = r.Header.Get("IFTTT-Service-Key")
	}

	defer r.Body.Close()

	if data, err := ioutil.ReadAll(r.Body); err != nil {
		return nil, err
	} else if len(data) > 0 {
		if decodeData, err := gabs.ParseJSON(data); err != nil {
			return nil, err
		} else {
			res.DecodedBody = decodeData
		}
	}

	switch {
	case strings.HasPrefix(r.RequestURI, "/ifttt/v1/user/info"):
		switch r.Method {
		case "GET":
			res.Type = UserInfoRequest
		}
	case strings.HasPrefix(r.RequestURI, "/ifttt/v1/status"):
		switch r.Method {
		case "GET":
			res.Type = ServiceStatus
		}
	case strings.HasPrefix(r.RequestURI, "/ifttt/v1/triggers"):
		switch {
		case urlRegexpTrigger.MatchString(r.RequestURI):
			switch r.Method {
			case "POST":
				match := urlRegexpTrigger.FindStringSubmatch(r.RequestURI)
				res.Slug = match[1]
				res.TriggerIdentity = res.DecodedBody.Path("trigger_identity").Data().(string)
				res.Type = TriggerFetch
			}
		case urlRegexpTriggerIdentity.MatchString(r.RequestURI):
			switch r.Method {
			case "DELETE":
				match := urlRegexpTriggerIdentity.FindStringSubmatch(r.RequestURI)
				res.Slug = match[1]
				res.TriggerIdentity = match[2]
				res.Type = TriggerDeleteNotify
			}
		case urlRegexpTriggerFieldOption.MatchString(r.RequestURI):
			switch r.Method {
			case "POST":
				match := urlRegexpTriggerFieldOption.FindStringSubmatch(r.RequestURI)
				res.Slug = match[1]
				res.FieldSlug = match[2]
				res.Type = TriggerDynamicOptions
			}
		case urlRegexpTriggerFieldOptionValidation.MatchString(r.RequestURI):
			switch r.Method {
			case "POST":
				match := urlRegexpTriggerFieldOptionValidation.FindStringSubmatch(r.RequestURI)
				res.Slug = match[1]
				res.FieldSlug = match[2]
				res.Type = TriggerDynamicValidation
			}
		case urlRegexpTriggerContextualValidation.MatchString(r.RequestURI):
			switch r.Method {
			case "POST":
				match := urlRegexpTriggerContextualValidation.FindStringSubmatch(r.RequestURI)
				res.Slug = match[1]
				res.Type = TriggerContextualValidation
			}
		}
	case strings.HasPrefix(r.RequestURI, "/ifttt/v1/actions"):
		switch {
		case urlRegexpAction.MatchString(r.RequestURI):
			switch r.Method {
			case "POST":
				match := urlRegexpAction.FindStringSubmatch(r.RequestURI)
				res.Slug = match[1]
				res.Type = ActionTrigger
			}
		case urlRegexpActionFieldOption.MatchString(r.RequestURI):
			switch r.Method {
			case "POST":
				match := urlRegexpActionFieldOption.FindStringSubmatch(r.RequestURI)
				res.Slug = match[1]
				res.FieldSlug = match[2]
				res.Type = ActionDynamicOptions
			}
		}
	}
	return res, nil
}
