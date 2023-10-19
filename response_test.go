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

func TestWithResponseStatusCodeAssertion(t *testing.T) {
	t.Run("response and asserted HTTP code match", func(t *testing.T) {
		assert.NoError(t, MoqResponse().Handle(WithResponseStatusCodeAssertion(http.StatusOK)))
	})
	t.Run("empty body. response and asserted HTTP code mismatch", func(t *testing.T) {
		assert.Equal(t, MoqResponse().Handle(WithResponseStatusCodeAssertion(http.StatusCreated)).Error(), "expected status code(s) '[201]', received '200'")
	})
	t.Run("non-empty body. response and asserted HTTP code mismatch ", func(t *testing.T) {
		assert.Equal(t, MoqResponse(func(response *Response) {
			response.Body = io.NopCloser(strings.NewReader("this is an error"))
		}).Handle(WithResponseStatusCodeAssertion(http.StatusCreated)).Error(), "this is an error")
	})
}

func TestWithResponseJSON(t *testing.T) {
	type testOK struct {
		Status string `json:","`
	}

	t.Run("body is JSON deserialized to given object", func(t *testing.T) {
		resultOK := &testOK{}
		err := MoqResponse(func(response *Response) {
			body, _ := json.Marshal(&testOK{Status: "ok"})
			response.Body = io.NopCloser(bytes.NewReader(body))
		}).Handle(
			WithResponseJSON(resultOK, http.StatusOK),
			WithResponseJSON(resultOK, http.StatusInternalServerError),
		)

		assert.NoError(t, err)
		assert.Equal(t, "ok", resultOK.Status)
	})

	t.Run("body is JSON deserialized to nil", func(t *testing.T) {
		var resultOK *testOK
		err := MoqResponse(func(response *Response) {
			body, _ := json.Marshal(&testOK{Status: "ok"})
			response.Body = io.NopCloser(bytes.NewReader(body))
		}).Handle(
			WithResponseJSON(resultOK),
		)

		assert.Error(t, err)
	})
}

func TestWithResponseXML(t *testing.T) {
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
			WithResponseXML(resultOK, http.StatusOK),
		)

		assert.NoError(t, err)
		assert.Equal(t, "github", resultOK.Name)
	})
}
