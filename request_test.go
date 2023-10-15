package rejester

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDry(t *testing.T) {

}

func TestDo(t *testing.T) {

}

func TestWithRequestOptions(t *testing.T) {

}

func TestWithRetryPolicy(t *testing.T) {

}

func TestWithTimeout(t *testing.T) {
	err := New().
		GET(context.Background(), "http://www.google.com:81").
		Do(WithTimeout(time.Second)).Err

	assert.Error(t, err)
}

func TestWithURL(t *testing.T) {
	request := New().
		GET(context.Background(), testURL)
	err := request.Dry(WithURL("https://test.no"))

	assert.NoError(t, err)
	assert.Equal(t, "https://test.no", request.URL.String())
}

func TestWithURLQuery(t *testing.T) {
	request := New().
		GET(context.Background(), testURL)
	err := request.Dry(WithURLQuery(map[string][]any{
		"id": {"123", 321},
	}))

	assert.NoError(t, err)
	assert.Equal(t, request.URL.String(), fmt.Sprintf("%s?id=123&id=321", testURL))
}

func TestWithBody(t *testing.T) {
	request := New().
		GET(context.Background(), testURL)
	err := request.Dry(WithBody(strings.NewReader("123")))

	assert.NoError(t, err)
	body, err := io.ReadAll(request.Body)
	assert.NoError(t, err)
	assert.Equal(t, "123", string(body))
}

func TestWithBodyXML(t *testing.T) {
	type TestXML struct {
		XMLName xml.Name `xml:"test"`
		Id      int      `xml:"id,attr"`
		Name    string   `xml:"name"`
		Origin  []string `xml:"origin"`
	}

	request := New().
		POST(context.Background(), testURL)

	err := request.Dry(WithBodyXML(&TestXML{
		Name: "github",
	}))

	assert.NoError(t, err)

	body, err := io.ReadAll(request.Body)
	assert.NoError(t, err)

	result := &TestXML{}
	err = xml.Unmarshal(body, result)
	assert.NoError(t, err)

	assert.Equal(t, "github", result.Name)
	assert.Equal(t, "application/xml", request.Header.Get("Content-Type"))
}

func TestWithBodyJSON(t *testing.T) {
	type TestJSON struct {
		Id int `json:"id"`
	}

	request := New().
		POST(context.Background(), testURL)

	err := request.Dry(WithBodyJSON(&TestJSON{
		Id: 123,
	}))

	assert.NoError(t, err)

	body, err := io.ReadAll(request.Body)
	assert.NoError(t, err)

	result := &TestJSON{}
	err = json.Unmarshal(body, result)
	assert.NoError(t, err)

	assert.Equal(t, 123, result.Id)
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
}

func TestWithBodyFormURLEncoded(t *testing.T) {
	request := New().
		POST(context.Background(), testURL)

	err := request.Dry(WithBodyFormURLEncoded(map[string][]string{
		"test": {"1", "3"},
	}))

	assert.NoError(t, err)
	body, err := io.ReadAll(request.Body)
	assert.NoError(t, err)
	assert.Equal(t, "test=1&test=3", string(body))
	assert.Equal(t, "application/x-www-form-urlencoded", request.Header.Get("Content-Type"))
}

func TestWithBodyFormData(t *testing.T) {
	request := New().
		POST(context.Background(), testURL)

	err := request.Dry(WithBodyFormData(map[string][]byte{
		"test": []byte("123"),
	}))

	mediatype, param, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	assert.NoError(t, err)
	reader := multipart.NewReader(request.Body, param["boundary"])
	form, err := reader.ReadForm(100)

	assert.NoError(t, err)
	assert.Equal(t, []string{"123"}, form.Value["test"])
	assert.Equal(t, "multipart/form-data", mediatype)
}

func TestWithAuthorizationBasic(t *testing.T) {
	request := New().POST(context.Background(), testURL)
	err := request.Dry(WithAuthorizationBasic("123", "321"))

	assert.NoError(t, err)
	assert.Equal(t, "Basic MTIzOjMyMQ==", request.Header.Get("Authorization"))
}

func TestWithAuthorizationBearer(t *testing.T) {
	request := New().POST(context.Background(), testURL)
	err := request.Dry(WithAuthorizationBearer(func(ctx context.Context) (string, error) {
		return "123", nil
	}))

	assert.NoError(t, err)
	assert.Equal(t, "Bearer 123", request.Header.Get("Authorization"))
}

func TestWithHeader(t *testing.T) {
	request := New().POST(context.Background(), testURL)
	err := request.Dry(WithHeader("X-TEST", 1))

	assert.NoError(t, err)
	assert.Equal(t, "1", request.Header.Get("X-TEST"))
}
