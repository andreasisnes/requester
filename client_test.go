package requester

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testURL = "https://test.com"

func TestNew(t *testing.T) {
	t.Run("URL being set", func(t *testing.T) {
		expected := "https://test.com"
		actual := New(WithBaseURL(expected)).url
		assert.Equal(t, expected, actual)
	})
	t.Run("HTTP client is being set", func(t *testing.T) {
		expected := &http.Client{Timeout: time.Hour}
		actual := New(WithClient(expected)).Client
		assert.Equal(t, expected, actual)
		assert.NotEqual(t, expected, http.DefaultClient)
	})
}

func TestDELETE(t *testing.T) {
	t.Run("HTTP method is DELETE", func(t *testing.T) {
		actual := New(WithBaseURL(testURL)).DELETE(context.Background()).Method
		assert.Equal(t, http.MethodDelete, actual)
	})
}

func TestPUT(t *testing.T) {
	t.Run("HTTP method is PUT", func(t *testing.T) {
		actual := New(WithBaseURL(testURL)).PUT(context.Background()).Method
		assert.Equal(t, http.MethodPut, actual)
	})
}

func TestGET(t *testing.T) {
	t.Run("HTTP method is GET", func(t *testing.T) {
		actual := New(WithBaseURL(testURL)).GET(context.Background()).Method
		assert.Equal(t, http.MethodGet, actual)
	})
}

func TestPOST(t *testing.T) {
	t.Run("HTTP method is POST", func(t *testing.T) {
		actual := New(WithBaseURL(testURL)).POST(context.Background()).Method
		assert.Equal(t, http.MethodPost, actual)
	})
}

func TestPATCH(t *testing.T) {
	t.Run("HTTP method is PATCH", func(t *testing.T) {
		actual := New(WithBaseURL(testURL)).PATCH(context.Background()).Method
		assert.Equal(t, http.MethodPatch, actual)
	})
}

func TestRequest(t *testing.T) {
	t.Run("URL base and routes is concatenated", func(t *testing.T) {
		actual := New(WithBaseURL(testURL)).Request(context.Background(), http.MethodGet, "1", "2")
		assert.Equal(t, fmt.Sprintf("%s/%s/%s", testURL, "1", "2"), actual.URL.String())
	})
	t.Run("URL routes is concatenated", func(t *testing.T) {
		actual := New().Request(context.Background(), http.MethodGet, testURL, "1", "2")
		assert.Equal(t, fmt.Sprintf("%s/%s/%s", testURL, "1", "2"), actual.URL.String())
	})
	t.Run("URL invalid is empty", func(t *testing.T) {
		actual := New().Request(context.Background(), http.MethodGet, "#").URL
		assert.Empty(t, actual.Scheme)
	})
	t.Run("Unknown HTTP verb return error", func(t *testing.T) {
		actual := New().Request(context.Background(), "INVALID HTTP VERB")
		assert.Error(t, actual.Error)
		assert.Nil(t, actual.Request)
	})
}
