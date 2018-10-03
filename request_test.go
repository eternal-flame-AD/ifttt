package ifttt

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func mockHeader(headers string, r *http.Request) {
	for _, head := range strings.Split(headers, "\n") {
		res := strings.SplitN(head, ": ", 2)
		if len(res) != 2 {
			continue
		}
		for res[0][0] == '\t' || res[0][0] == ' ' {
			res[0] = res[0][1:]
		}
		r.Header.Set(res[0], res[1])
	}
}

func TestRequestParse(t *testing.T) {

	Convey("Test Request Parse", t, func() {
		Convey("Mock Trigger Poll", func() {
			req := httptest.NewRequest("POST", "/ifttt/v1/triggers/test_trigger", bytes.NewBufferString(`
			{
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
			Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
			Accept: application/json
			Accept-Charset: utf-8
			Accept-Encoding: gzip, deflate
			Content-Type: application/json
			X-Request-ID: 7f7cd9e0d8154531bbf36da8fe24b449`, req)

			res, err := parseRequest(req)

			So(err, ShouldBeNil)

			So(*res, ShouldResemble, Request{
				Authenticated:   true,
				UserAccessToken: "b29a71b4c58c22af116578a6be6402d2",
				RequestUUID:     "7f7cd9e0d8154531bbf36da8fe24b449",
				Slug:            "test_trigger",
				FieldSlug:       "",
				TriggerIdentity: "92429d82a41e93048",
				DecodedBody:     res.DecodedBody,
				Type:            TriggerFetch,
				RawRequest:      res.RawRequest,
			})
		})

		Convey("Mock Trigger Identity Delete", func() {
			req := httptest.NewRequest("DELETE", "/ifttt/v1/triggers/new_photo_in_album_with_hashtag/trigger_identity/92429d82a41e93048", bytes.NewBufferString(""))

			mockHeader(`Host: api.example-service.com
			Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
			Accept: application/json
			Accept-Charset: utf-8
			Accept-Encoding: gzip, deflate
			Content-Type: application/json
			X-Request-ID: 7f7cd9e0d8154531bbf36da8fe24b449`, req)

			res, err := parseRequest(req)

			So(err, ShouldBeNil)

			So(*res, ShouldResemble, Request{
				Authenticated:   true,
				UserAccessToken: "b29a71b4c58c22af116578a6be6402d2",
				RequestUUID:     "7f7cd9e0d8154531bbf36da8fe24b449",
				Slug:            "new_photo_in_album_with_hashtag",
				FieldSlug:       "",
				TriggerIdentity: "92429d82a41e93048",
				DecodedBody:     res.DecodedBody,
				Type:            TriggerDeleteNotify,
				RawRequest:      res.RawRequest,
			})
		})

		Convey("Mock Trigger Dynamic Field Request", func() {
			req := httptest.NewRequest("POST", "/ifttt/v1/triggers/new_photo_in_album_with_hashtag/fields/album_name/options", bytes.NewBufferString("{}"))

			mockHeader(`Host: api.example-service.com
			Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
			Accept: application/json
			Accept-Charset: utf-8
			Accept-Encoding: gzip, deflate
			X-Request-ID: 37ccb881af5542fe8c5534e9744b6116`, req)

			res, err := parseRequest(req)

			So(err, ShouldBeNil)

			So(*res, ShouldResemble, Request{
				Authenticated:   true,
				UserAccessToken: "b29a71b4c58c22af116578a6be6402d2",
				RequestUUID:     "37ccb881af5542fe8c5534e9744b6116",
				Slug:            "new_photo_in_album_with_hashtag",
				FieldSlug:       "album_name",
				TriggerIdentity: "",
				DecodedBody:     res.DecodedBody,
				Type:            TriggerDynamicOptions,
				RawRequest:      res.RawRequest,
			})
		})

		Convey("Mock Trigger Dynamic Validation", func() {
			req := httptest.NewRequest("POST", "/ifttt/v1/triggers/new_photo_in_album_with_hashtag/fields/album_name/validate", bytes.NewBufferString(`{
				"value": "Street Art"
			}`))

			mockHeader(`Host: api.example-service.com
			Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
			Accept: application/json
			Accept-Charset: utf-8
			Accept-Encoding: gzip, deflate
			Content-Type: application/json
			X-Request-ID: b959f481ef4f4a8ab0ec414f58991674`, req)

			res, err := parseRequest(req)

			So(err, ShouldBeNil)

			So(*res, ShouldResemble, Request{
				Authenticated:   true,
				UserAccessToken: "b29a71b4c58c22af116578a6be6402d2",
				RequestUUID:     "b959f481ef4f4a8ab0ec414f58991674",
				Slug:            "new_photo_in_album_with_hashtag",
				FieldSlug:       "album_name",
				TriggerIdentity: "",
				DecodedBody:     res.DecodedBody,
				Type:            TriggerDynamicValidation,
				RawRequest:      res.RawRequest,
			})
		})

		Convey("Mock Trigger Contextual Validation", func() {
			req := httptest.NewRequest("POST", "/ifttt/v1/triggers/new_comment_on_card/validate", bytes.NewBufferString(`{
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

			res, err := parseRequest(req)

			So(err, ShouldBeNil)

			So(*res, ShouldResemble, Request{
				Authenticated:   true,
				UserAccessToken: "b29a71b4c58c22af116578a6be6402d2",
				RequestUUID:     "b959f481ef4f4a8ab0ec414f58991674",
				Slug:            "new_comment_on_card",
				FieldSlug:       "",
				TriggerIdentity: "",
				DecodedBody:     res.DecodedBody,
				Type:            TriggerContextualValidation,
				RawRequest:      res.RawRequest,
			})
		})

		Convey("Mock Action Trigger", func() {
			req := httptest.NewRequest("POST", "/ifttt/v1/actions/new_status_update", bytes.NewBufferString(`{
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
			  }
			`))
			mockHeader(`Host: api.example-service.com
			Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
			Accept: application/json
			Accept-Charset: utf-8
			Accept-Encoding: gzip, deflate
			Content-Type: application/json
			X-Request-ID: 1d21c3cd2ed8441ea269dd554d2c8e54`, req)

			res, err := parseRequest(req)

			So(err, ShouldBeNil)

			So(*res, ShouldResemble, Request{
				Authenticated:   true,
				UserAccessToken: "b29a71b4c58c22af116578a6be6402d2",
				RequestUUID:     "1d21c3cd2ed8441ea269dd554d2c8e54",
				Slug:            "new_status_update",
				FieldSlug:       "",
				TriggerIdentity: "",
				DecodedBody:     res.DecodedBody,
				Type:            ActionTrigger,
				RawRequest:      res.RawRequest,
			})
		})

		Convey("Mock Action Dynamic Option", func() {
			req := httptest.NewRequest("POST", "/ifttt/v1/actions/post_photo_to_album/fields/album_name/options", bytes.NewBufferString("{}"))

			mockHeader(`Host: api.example-service.com
			Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
			Accept: application/json
			Accept-Charset: utf-8
			Accept-Encoding: gzip, deflate
			X-Request-ID: 9f99e73452cd40198cb6ce9c1cde83d6`, req)

			res, err := parseRequest(req)

			So(err, ShouldBeNil)

			So(*res, ShouldResemble, Request{
				Authenticated:   true,
				UserAccessToken: "b29a71b4c58c22af116578a6be6402d2",
				RequestUUID:     "9f99e73452cd40198cb6ce9c1cde83d6",
				Slug:            "post_photo_to_album",
				FieldSlug:       "album_name",
				TriggerIdentity: "",
				DecodedBody:     res.DecodedBody,
				Type:            ActionDynamicOptions,
				RawRequest:      res.RawRequest,
			})
		})

		Convey("Mock User Info", func() {
			req := httptest.NewRequest("GET", "/ifttt/v1/user/info", bytes.NewBufferString(""))

			mockHeader(`Host: api.example-service.com
			Authorization: Bearer b29a71b4c58c22af116578a6be6402d2
			Accept: application/json
			Accept-Charset: utf-8
			Accept-Encoding: gzip, deflate
			X-Request-ID: 434d757081c94013b1b28f2087d28a98`, req)

			res, err := parseRequest(req)

			So(err, ShouldBeNil)

			So(*res, ShouldResemble, Request{
				Authenticated:   true,
				UserAccessToken: "b29a71b4c58c22af116578a6be6402d2",
				RequestUUID:     "434d757081c94013b1b28f2087d28a98",
				Slug:            "",
				FieldSlug:       "",
				TriggerIdentity: "",
				DecodedBody:     res.DecodedBody,
				Type:            UserInfoRequest,
				RawRequest:      res.RawRequest,
			})
		})

		Convey("Mock Service Status", func() {
			req := httptest.NewRequest("GET", "/ifttt/v1/status", bytes.NewBufferString(""))

			mockHeader(`Host api.example-service.com
			IFTTT-Service-Key: vFRqPGZBmZjB8JPp3mBFqOdt
			Accept: application/json
			Accept-Charset: utf-8
			Accept-Encoding: gzip, deflate
			X-Request-ID: 0715f98e65f749aba2fc243eac1e3c09`, req)

			res, err := parseRequest(req)

			So(err, ShouldBeNil)

			So(*res, ShouldResemble, Request{
				Authenticated:   false,
				UserAccessToken: "vFRqPGZBmZjB8JPp3mBFqOdt",
				RequestUUID:     "0715f98e65f749aba2fc243eac1e3c09",
				Slug:            "",
				FieldSlug:       "",
				TriggerIdentity: "",
				DecodedBody:     res.DecodedBody,
				Type:            ServiceStatus,
				RawRequest:      res.RawRequest,
			})
		})
	})

}
