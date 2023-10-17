package requester

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func MoqResponse(opts ...func(response *Response)) *Response {
	response := &Response{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
		},
	}

	for _, opt := range opts {
		opt(response)
	}

	return response
}

func TestWithStatusCodeAssertion(t *testing.T) {
	t.Run("response and asserted HTTP code match", func(t *testing.T) {
		assert.NoError(t, MoqResponse().Handle(WithStatusCodeAssertion(http.StatusOK)))
	})
	t.Run("empty body. response and asserted HTTP code mismatch", func(t *testing.T) {
		assert.Equal(t, MoqResponse().Handle(WithStatusCodeAssertion(http.StatusCreated)).Error(), "expected status code(s) '[201]', received '200'")
	})
	t.Run("non-empty body. response and asserted HTTP code mismatch ", func(t *testing.T) {
		assert.Equal(t, MoqResponse(func(response *Response) {
			response.Body = io.NopCloser(strings.NewReader("this is an error"))
		}).Handle(WithStatusCodeAssertion(http.StatusCreated)).Error(), "this is an error")
	})
}

func TestWithUnmarshalJSON(t *testing.T) {
	type testOK struct {
		Status string `json:","`
	}

	t.Run("body is JSON deserialized to given object", func(t *testing.T) {
		resultOK := &testOK{}
		err := MoqResponse(func(response *Response) {
			body, _ := json.Marshal(&testOK{Status: "ok"})
			response.Body = io.NopCloser(bytes.NewReader(body))
		}).Handle(
			WithUnmarshalJSON(resultOK, http.StatusOK),
			WithUnmarshalJSON(resultOK, http.StatusInternalServerError),
		)

		assert.NoError(t, err)
		assert.Equal(t, "ok", resultOK.Status)
	})
}

func TestWithUnmarshalXML(t *testing.T) {
	type testOK struct {
		XMLName xml.Name `xml:"test"`
		Id      int      `xml:"id,attr"`
		Name    string   `xml:"name"`
		Origin  []string `xml:"origin"`
	}

	t.Run("body is XML deserialized to given object", func(t *testing.T) {
		resultOK := &testOK{}
		err := MoqResponse(func(response *Response) {
			body, _ := xml.Marshal(&testOK{Id: 2, Name: "github"})
			response.Body = io.NopCloser(bytes.NewReader(body))
		}).Handle(
			WithUnmarshalXML(resultOK, http.StatusOK),
		)

		assert.NoError(t, err)
		assert.Equal(t, "github", resultOK.Name)
	})
}
