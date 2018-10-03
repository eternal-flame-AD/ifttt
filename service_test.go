package ifttt

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Jeffail/gabs"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	health bool
)

type testAction struct {
	ShouldPanic bool
}

func (c testAction) Options(req *Request) (*DynamicOption, error) {
	if c.ShouldPanic {
		panic("I am mad!")
	}
	if req.UserAccessToken == "realsecrettoken" {
		if req.FieldSlug == "test_field" {
			option := new(DynamicOption)
			suboption := new(DynamicOption)

			option.AddString("foo", "123")
			suboption.AddString("baz", "456")
			suboption.AddString("bar", "789")
			option.AddCategory("bar", suboption)
			return option, nil
		} else {
			return nil, errors.New("Invalid field")
		}

	}
	if req.UserAccessToken == "wrongtoken" {
		return nil, ErrorInvalidToken
	}
	return nil, errors.New("Unknown error")
}

type testTrigger struct {
}

func (c testTrigger) RealTime() bool {
	return true
}

func (c testTrigger) Poll(req *TriggerPollRequest, r *Request) (TriggerEventCollection, error) {
	res := make(TriggerEventCollection, 0)

	evt1 := TriggerEvent{
		Slugs: req.TriggerFields,
		Meta: TriggerEventMeta{
			r.RequestUUID,
			time.Unix(10000, 0),
		},
	}

	res = append(res, evt1)

	return res, nil
}

func (c testTrigger) Options(req *Request) (*DynamicOption, error) {
	if req.UserAccessToken == "realsecrettoken" {
		if req.FieldSlug == "test_field" {
			option := new(DynamicOption)
			suboption := new(DynamicOption)

			option.AddString("foo", "123")
			suboption.AddString("baz", "456")
			suboption.AddString("bar", "789")
			option.AddCategory("bar", suboption)
			return option, nil
		} else {
			return nil, errors.New("Invalid field")
		}

	}
	if req.UserAccessToken == "wrongtoken" {
		return nil, ErrorInvalidToken
	}
	return nil, errors.New("Unknown error")
}

func (c testTrigger) ValidateField(fieldslug string, value string, req *Request) error {
	if fieldslug == "foo" && value == "bar" {
		return nil
	} else {
		return errors.New("Invalid combination")
	}
}

func (c testTrigger) ValidateContext(values map[string]string, req *Request) (map[string]error, error) {
	res := make(map[string]error)

	for key, val := range values {
		if val == "wrong" {
			res[key] = errors.New("Invalid value")
		} else {
			res[key] = nil
		}
	}
	return res, nil
}

func (c testTrigger) RemoveIdentity(triggerid string) error {
	return nil
}

func (c testAction) Handle(r *ActionHandleRequest, req *Request) (res *ActionResult, skip bool, err error) {
	if c.ShouldPanic {
		panic("I am mad!")
	}
	So(r.User, ShouldContainKey, "timezone")
	if req.UserAccessToken == "realsecrettoken" {
		return &ActionResult{
			"123",
			r.ActionFields["URL"],
		}, false, nil
	}
	if req.UserAccessToken == "wrongtoken" {
		return nil, true, ErrorInvalidToken
	}
	return nil, false, errors.New("Unknown error")
}

func TestService(t *testing.T) {

	Convey("Test Mock Service", t, func() {
		service := &Service{
			ServiceKey: "abcdef",
			Healthy: func() bool {
				return health
			},
			UserInfo: func(req *Request) (*UserInfo, error) {
				if req.UserAccessToken == "realsecrettoken" {
					return &UserInfo{
						"John",
						"1",
						"https://www.example.com",
					}, nil
				}
				if req.UserAccessToken == "wrongtoken" {
					return nil, ErrorInvalidToken
				}
				return nil, errors.New("Unknown error")
			},
		}
		service.EnableDebug()

		Convey("Test Health Check", func() {

			testthis := func(ok bool) {

				health = ok
				req := httptest.NewRequest("GET", "/ifttt/v1/status", bytes.NewBufferString(""))

				mockHeader(`Host api.example-service.com
				IFTTT-Service-Key: vFRqPGZBmZjB8JPp3mBFqOdt
				Accept: application/json
				Accept-Charset: utf-8
				Accept-Encoding: gzip, deflate
				X-Request-ID: 0715f98e65f749aba2fc243eac1e3c09`, req)

				res := httptest.NewRecorder()

				service.ServeHTTP(res, req)

				if ok {
					So(res.Code, ShouldEqual, 200)
				} else {
					So(res.Code, ShouldEqual, 503)
				}
			}

			testthis(true)
			testthis(false)

		})

		Convey("Test User Info", func() {

			req := httptest.NewRequest("GET", "/ifttt/v1/user/info", bytes.NewBufferString(""))
			mockHeader(`Host: api.example-service.com
			Authorization: Bearer Unknown
			Accept: application/json
			Accept-Charset: utf-8
			Accept-Encoding: gzip, deflate
			X-Request-ID: 434d757081c94013b1b28f2087d28a98`, req)

			res := httptest.NewRecorder()
			service.ServeHTTP(res, req)
			resbytes, _ := ioutil.ReadAll(res.Body)
			So(res.Code, ShouldEqual, 500)
			So(jsonEqual(resbytes, marshalError(errors.New("Unknown error"), false)), ShouldEqual, true)

			mockHeader(`Host: api.example-service.com
			Authorization: Bearer wrongtoken
			Accept: application/json
			Accept-Charset: utf-8
			Accept-Encoding: gzip, deflate
			X-Request-ID: 434d757081c94013b1b28f2087d28a98`, req)
			res = httptest.NewRecorder()
			service.ServeHTTP(res, req)
			So(res.Code, ShouldEqual, 401)

			mockHeader(`Host: api.example-service.com
			Authorization: Bearer realsecrettoken
			Accept: application/json
			Accept-Charset: utf-8
			Accept-Encoding: gzip, deflate
			X-Request-ID: 434d757081c94013b1b28f2087d28a98`, req)
			res = httptest.NewRecorder()
			service.ServeHTTP(res, req)
			resbytes, _ = ioutil.ReadAll(res.Body)
			So(res.Code, ShouldEqual, 200)
			So(jsonEqual(resbytes, []byte(`{"data":{"id":"1","name":"John","url":"https://www.example.com"}}`)), ShouldEqual, true)

		})

		Convey("Test Actions", func() {

			Convey("Test Action Panic", func() {
				actionpanic := testAction{true}
				service.RegisterAction("action_panic", actionpanic)

				req := httptest.NewRequest("POST", "/ifttt/v1/actions/action_panic", bytes.NewBufferString(`{
					"actionFields": {
					  "title": "New Banksy photo!",
					  "body": "Check out a new Bansky photo: http://example.com/images/125"
					},
					"ifttt_source": {
					  "id": "2",
					  "url": "https://ifttt.com/myrecipes/personal/2"
					},
					"user": {
					  "timezone": "Pacific Time (US & Canada)"
					}
				}`))

				mockHeader(`Host: api.example-service.com
				Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
				Accept: application/json
				Accept-Charset: utf-8
				Accept-Encoding: gzip, deflate
				Content-Type: application/json
				X-Request-ID: 1d21c3cd2ed8441ea269dd554d2c8e54`, req)

				res := httptest.NewRecorder()

				service.ServeHTTP(res, req)

				resbytes, _ := ioutil.ReadAll(res.Body)
				So(res.Code, ShouldEqual, 500)
				So(resbytes, ShouldResemble, marshalError(ErrorPanicDuringProcess, false))

			})

			action := testAction{}

			service.RegisterAction("test_action", action)
			Convey("Test Action Trigger", func() {
				req := httptest.NewRequest("POST", "/ifttt/v1/actions/test_action", bytes.NewBufferString(`{
					"actionFields": {
					  "title": "New Banksy photo!",
					  "body": "Check out a new Bansky photo: http://example.com/images/125"
					},
					"ifttt_source": {
					  "id": "2",
					  "url": "https://ifttt.com/myrecipes/personal/2"
					},
					"user": {
					  "timezone": "Pacific Time (US & Canada)"
					}
				}`))

				mockHeader(`Host: api.example-service.com
				Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
				Accept: application/json
				Accept-Charset: utf-8
				Accept-Encoding: gzip, deflate
				Content-Type: application/json
				X-Request-ID: 1d21c3cd2ed8441ea269dd554d2c8e54`, req)

				res := httptest.NewRecorder()

				service.ServeHTTP(res, req)

				resbytes, _ := ioutil.ReadAll(res.Body)
				So(res.Code, ShouldEqual, 400)
				So(jsonEqual(resbytes, marshalError(errors.New("Unknown error"), false)), ShouldEqual, true)

				req = httptest.NewRequest("POST", "/ifttt/v1/actions/test_action", bytes.NewBufferString(`{
					"actionFields": {
					  "title": "New Banksy photo!",
					  "body": "Check out a new Bansky photo: http://example.com/images/125"
					},
					"ifttt_source": {
					  "id": "2",
					  "url": "https://ifttt.com/myrecipes/personal/2"
					},
					"user": {
					  "timezone": "Pacific Time (US & Canada)"
					}
				}`))

				mockHeader(`Host: api.example-service.com
				Authorization: Bearer wrongtoken
				Accept: application/json
				Accept-Charset: utf-8
				Accept-Encoding: gzip, deflate
				Content-Type: application/json
				X-Request-ID: 1d21c3cd2ed8441ea269dd554d2c8e54`, req)

				res = httptest.NewRecorder()

				service.ServeHTTP(res, req)

				So(res.Code, ShouldEqual, 401)

				req = httptest.NewRequest("POST", "/ifttt/v1/actions/test_action", bytes.NewBufferString(`{
					"actionFields": {
					  "title": "New Banksy photo!",
					  "URL": "http://example.com/images/125"
					},
					"ifttt_source": {
					  "id": "2",
					  "url": "https://ifttt.com/myrecipes/personal/2"
					},
					"user": {
					  "timezone": "Pacific Time (US & Canada)"
					}
				}`))

				mockHeader(`Host: api.example-service.com
				Authorization: Bearer realsecrettoken
				Accept: application/json
				Accept-Charset: utf-8
				Accept-Encoding: gzip, deflate
				Content-Type: application/json
				X-Request-ID: 1d21c3cd2ed8441ea269dd554d2c8e54`, req)

				res = httptest.NewRecorder()

				service.ServeHTTP(res, req)

				resbytes, _ = ioutil.ReadAll(res.Body)
				So(res.Code, ShouldEqual, 200)
				So(jsonEqual(resbytes, []byte(`{"data":[{"id":"123","url":"http://example.com/images/125"}]}`)), ShouldEqual, true)

			})

			Convey("Test Action Dynamic Options", func() {

				req := httptest.NewRequest("POST", "/ifttt/v1/actions/test_action/fields/test_field/options", bytes.NewBufferString("{}"))
				mockHeader(`Host: api.example-service.com
				Authorization: Bearer realsecrettoken
				Accept: application/json
				Accept-Charset: utf-8
				Accept-Encoding: gzip, deflate
				X-Request-ID: 9f99e73452cd40198cb6ce9c1cde83d6`, req)
				res := httptest.NewRecorder()

				service.ServeHTTP(res, req)

				resbytes, _ := ioutil.ReadAll(res.Body)
				So(res.Code, ShouldEqual, 200)

				resobj, err := gabs.ParseJSON(resbytes)
				So(err, ShouldBeNil)
				size, err := resobj.ArrayCount("data")
				So(err, ShouldBeNil)
				So(size, ShouldEqual, 2)
				objfoo, err := resobj.ArrayElement(0, "data")
				So(err, ShouldBeNil)
				objagg, err := resobj.ArrayElement(1, "data")
				So(err, ShouldBeNil)
				if objfoo.Exists("values") {
					objagg, objfoo = objfoo, objagg
				}
				So(objfoo.S("label").Data(), ShouldEqual, "foo")
				So(objfoo.S("value").Data(), ShouldEqual, "123")

				So(objagg.S("label").Data(), ShouldEqual, "bar")
				aggcount, err := objagg.ArrayCount("values")
				So(err, ShouldBeNil)
				So(aggcount, ShouldEqual, 2)
				obj1 := objagg.S("values").Data().([]interface{})[0]
				So([]string{"baz", "bar"}, ShouldContain, obj1.(map[string]interface{})["label"].(string))
				So([]string{"456", "789"}, ShouldContain, obj1.(map[string]interface{})["value"].(string))

			})

		})

		Convey("Test Triggers", func() {

			trigger := testTrigger{}

			service.RegisterTrigger("test_trigger", trigger)

			Convey("Test Trigger Fetch", func() {
				req := httptest.NewRequest("POST", "/ifttt/v1/triggers/test_trigger", bytes.NewBufferString(`{
					"trigger_identity": "92429d82a41e93048",
					"triggerFields": {
					  "album_name": "Street Art",
					  "hashtag": "banksy"
					},
					"ifttt_source": {
					  "id": "2",
					  "url": "https://ifttt.com/myrecipes/personal/2"
					},
					"user": {
					  "timezone": "Pacific Time (US & Canada)"
					}
				}`))
				mockHeader(`Host: api.example-service.com
				Authorization: Bearer realsecrettoken
				Accept: application/json
				Accept-Charset: utf-8
				Accept-Encoding: gzip, deflate
				Content-Type: application/json
				X-Request-ID: 7f7cd9e0d8154531bbf36da8fe24b449`, req)

				res := httptest.NewRecorder()

				service.ServeHTTP(res, req)
				resbytes, _ := ioutil.ReadAll(res.Body)
				So(res.Code, ShouldEqual, 200)
				So(jsonEqual(resbytes, []byte(`{"data":[{"album_name":"Street Art","hashtag":"banksy","meta":{"id":"7f7cd9e0d8154531bbf36da8fe24b449","timestamp":10000}}]}`)), ShouldEqual, true)
				So(res.Header().Get("X-IFTTT-REALTIME"), ShouldEqual, "1")
			})

			Convey("Test Trigger Dynamic Options", func() {
				req := httptest.NewRequest("POST", "/ifttt/v1/triggers/test_trigger/fields/test_field/options", bytes.NewBufferString("{}"))
				mockHeader(`Host: api.example-service.com
				Authorization: Bearer realsecrettoken
				Accept: application/json
				Accept-Charset: utf-8
				Accept-Encoding: gzip, deflate
				X-Request-ID: 9f99e73452cd40198cb6ce9c1cde83d6`, req)
				res := httptest.NewRecorder()

				service.ServeHTTP(res, req)

				resbytes, _ := ioutil.ReadAll(res.Body)
				So(res.Code, ShouldEqual, 200)

				resobj, err := gabs.ParseJSON(resbytes)
				So(err, ShouldBeNil)
				size, err := resobj.ArrayCount("data")
				So(err, ShouldBeNil)
				So(size, ShouldEqual, 2)
				objfoo, err := resobj.ArrayElement(0, "data")
				So(err, ShouldBeNil)
				objagg, err := resobj.ArrayElement(1, "data")
				So(err, ShouldBeNil)
				if objfoo.Exists("values") {
					objagg, objfoo = objfoo, objagg
				}
				So(objfoo.S("label").Data(), ShouldEqual, "foo")
				So(objfoo.S("value").Data(), ShouldEqual, "123")
				So(objagg.S("label").Data(), ShouldEqual, "bar")
				aggcount, err := objagg.ArrayCount("values")
				So(err, ShouldBeNil)
				So(aggcount, ShouldEqual, 2)
				obj1 := objagg.S("values").Data().([]interface{})[0]
				So([]string{"baz", "bar"}, ShouldContain, obj1.(map[string]interface{})["label"].(string))
				So([]string{"456", "789"}, ShouldContain, obj1.(map[string]interface{})["value"].(string))
			})

			Convey("Test Trigger Field Validation", func() {
				req := httptest.NewRequest("POST", "/ifttt/v1/triggers/test_trigger/fields/foo/validate", bytes.NewBufferString(`{
					"value": "Street Art"
				}`))
				mockHeader(`Host: api.example-service.com
				Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
				Accept: application/json
				Accept-Charset: utf-8
				Accept-Encoding: gzip, deflate
				Content-Type: application/json
				X-Request-ID: b959f481ef4f4a8ab0ec414f58991674`, req)

				res := httptest.NewRecorder()
				service.ServeHTTP(res, req)

				resbytes, _ := ioutil.ReadAll(res.Body)
				So(res.Code, ShouldEqual, 200)

				So(jsonEqual(resbytes, []byte(`{
					"data":  {
					  "message": "Invalid combination",
					  "valid": false
					}
				}`)), ShouldEqual, true)

				req = httptest.NewRequest("POST", "/ifttt/v1/triggers/test_trigger/fields/foo/validate", bytes.NewBufferString(`{
					"value": "bar"
				}`))
				mockHeader(`Host: api.example-service.com
				Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
				Accept: application/json
				Accept-Charset: utf-8
				Accept-Encoding: gzip, deflate
				Content-Type: application/json
				X-Request-ID: b959f481ef4f4a8ab0ec414f58991674`, req)

				res = httptest.NewRecorder()
				service.ServeHTTP(res, req)

				resbytes, _ = ioutil.ReadAll(res.Body)
				So(res.Code, ShouldEqual, 200)

				So(jsonEqual(resbytes, []byte(`{
					"data":  {
					  "valid": true
					}
				}`)), ShouldEqual, true)

			})

			Convey("Test Trigger Contextual Validation", func() {
				req := httptest.NewRequest("POST", "/ifttt/v1/triggers/test_trigger/validate", bytes.NewBufferString(`{
				  "values": {
					"board": "New Features",
					"card": "Potential Ideas"
				  }
				}`))
				mockHeader(`Host: api.example-service.com
				Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
				Accept: application/json
				Accept-Charset: utf-8
				Accept-Encoding: gzip, deflate
				Content-Type: application/json
				X-Request-ID: b959f481ef4f4a8ab0ec414f58991674`, req)

				res := httptest.NewRecorder()
				service.ServeHTTP(res, req)

				resbytes, _ := ioutil.ReadAll(res.Body)
				So(res.Code, ShouldEqual, 200)
				So(jsonEqual(resbytes, []byte(`{"data":{"board":{"valid":true},"card":{"valid":true}}}`)), ShouldEqual, true)

				req = httptest.NewRequest("POST", "/ifttt/v1/triggers/test_trigger/validate", bytes.NewBufferString(`{
					"values": {
					  "board": "wrong",
					  "card": "Potential Ideas"
					}
				  }`))
				mockHeader(`Host: api.example-service.com
				  Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
				  Accept: application/json
				  Accept-Charset: utf-8
				  Accept-Encoding: gzip, deflate
				  Content-Type: application/json
				  X-Request-ID: b959f481ef4f4a8ab0ec414f58991674`, req)

				res = httptest.NewRecorder()
				service.ServeHTTP(res, req)

				resbytes, _ = ioutil.ReadAll(res.Body)
				So(res.Code, ShouldEqual, 200)
				So(jsonEqual(resbytes, []byte(`{"data":{"board":{"valid":false,"message":"Invalid value"},"card":{"valid":true}}}`)), ShouldEqual, true)
			})

			Convey("Test Trigger Delete Notify", func() {
				req := httptest.NewRequest("DELETE", "/ifttt/v1/triggers/test_trigger/trigger_identity/92429d82a41e93048", bytes.NewBuffer([]byte{}))
				mockHeader(`Host: api.example-service.com
				Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
				Accept: application/json
				Accept-Charset: utf-8
				Accept-Encoding: gzip, deflate
				Content-Type: application/json
				X-Request-ID: 7f7cd9e0d8154531bbf36da8fe24b449`, req)

				res := httptest.NewRecorder()
				service.ServeHTTP(res, req)

				So(res.Code, ShouldEqual, 200)
			})

		})
	})

}
