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

// RequestType enumerates IFTTT request types
type RequestType int

const (
	// Unknown unknown request types
	Unknown RequestType = iota
	// TriggerFetch polls triggers for update information
	// https://platform.ifttt.com/docs/api_reference#triggers
	TriggerFetch
	// TriggerDeleteNotify notifies the trigger service that a trigger identifies by trigger identity has been removed and the server can stop making notifications regarding this trigger
	// https://platform.ifttt.com/docs/api_reference#trigger-identity
	TriggerDeleteNotify
	// TriggerDynamicOptions requests dynamic field options from a trigger
	// https://platform.ifttt.com/docs/api_reference#trigger-field-dynamic-options
	TriggerDynamicOptions
	// TriggerDynamicValidation requests the validation of a single field of a trigger
	// https://platform.ifttt.com/docs/api_reference#trigger-field-dynamic-validation
	TriggerDynamicValidation
	// TriggerContextualValidation requests the validation of a combination of the trigger fields, you need to contact IFTTT to have this enabled
	// https://platform.ifttt.com/docs/api_reference#trigger-field-contextual-validation
	TriggerContextualValidation
	// ActionTrigger triggers an action
	// https://platform.ifttt.com/docs/api_reference#actions
	ActionTrigger
	// ActionDynamicOptions requests dynamic field options from an action
	// https://platform.ifttt.com/docs/api_reference#action-fields
	ActionDynamicOptions
	// ServiceStatus requests for the availability of the service
	// https://platform.ifttt.com/docs/api_reference#service-status
	ServiceStatus
	// UserInfoRequest requests for the information of a user which will be displayed to the user on his service page
	// https://platform.ifttt.com/docs/api_reference#user-information
	UserInfoRequest
)

// Request represents a parsed request from IFTTT
type Request struct {
	// Authenticated whether the request was carrying an OAuth token or not
	Authenticated bool
	// UserAccessToken if Authenticated is true, this is the token provided by IFTTT, is not, this is set to the claimed Service Key from IFTTT
	UserAccessToken string
	// RequestUUID is a unique string which identifies requests for debugging purposes
	RequestUUID string
	// Slug the slug name of the action/trigger, you generally dont need to check this as requests are automatically routed by the package to their endpoints.
	Slug string
	// FieldSlug the field slug requested by the request, empty string in non field-related requests
	FieldSlug string
	// TriggerIdentity the identification of the trigger which initiated this request, empty string if not available
	TriggerIdentity string
	// DecodedBody the decoded JSON body of the request, you can extract data manually if it was not exposed by this package
	DecodedBody *gabs.Container
	// Type the type of the request, you generally dont need to check this as requests are automatically routed by the package to their endpoints.
	Type RequestType
	// RawRequest the raw HTTP request from IFTTT
	RawRequest *http.Request
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
