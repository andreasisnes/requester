package requester

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// FallbackPolicy specifies the cooldown strategy for failed request
// issuing an identical request of a failing request
type FallbackPolicy int

const (
	// FallbackPolicyLinear waits for issuing a new request by
	// given duration multiplied by the attempt.
	FallbackPolicyLinear FallbackPolicy = iota
	// FallbackPolicyExponential waits for issuing a new request by
	// given attempt multiplied with itself and attempt.
	FallbackPolicyExponential
)

// RequestOption callback signature for modifying request
type RequestOption func(request *Request) (err error)

// Request is a wrapper around the standard http.Request providing additional features.
type Request struct {
	// Request is the underlying standard HTTP request being made.
	*http.Request

	// Client is the HTTP client used to perform the request.
	*http.Client

	// Error stores any errors generated when creating the request.
	Error error

	// Retries specifies the number of times the request will be retried in case of failure.
	Retries int

	// FallbackDuration is the duration to wait before attempting the request again.
	FallbackDuration time.Duration

	// FallbackPolicy represents the policy used for fallback requests.
	FallbackPolicy FallbackPolicy

	// FallbackStatusCodes contains a list of HTTP status codes that will
	// trigger a new request.
	FallbackStatusCodes []int
}

// Dry performs a dry run of the request without actually executing it.
func (r *Request) Dry(opts ...RequestOption) (err error) {
	if r.Error != nil {
		return r.Error
	}

	for _, o := range opts {
		err = errors.Join(r.Error, o(r))
	}

	return err
}

// Do executes the request.
func (r *Request) Do(opts ...RequestOption) *Response {
	if r.Error != nil || r.Request == nil {
		return &Response{Response: &http.Response{}, Err: r.Error}
	}

	errs := []error{}
	for _, o := range opts {
		errs = append(errs, o(r))
	}

	response, err := r.sender(0, nil, []error{})
	errs = append(errs, err...)

	return &Response{response, errors.Join(errs...)}
}

func (r *Request) sender(attempt int, response *http.Response, errs []error) (*http.Response, []error) {
	if 0 < attempt {
		if attempt >= r.Retries {
			return response, errs
		}

		switch r.FallbackPolicy {
		case FallbackPolicyExponential:
			r.wait(r.FallbackDuration * (time.Duration(attempt * attempt)))
		default:
			r.wait(r.FallbackDuration * time.Duration(attempt))
		}
	}

	attempt++
	response, err := r.Client.Do(r.Request)
	if err != nil {
		return r.sender(attempt, response, append(errs, err))
	}

	for _, statusCode := range r.FallbackStatusCodes {
		if statusCode == response.StatusCode {
			return r.sender(attempt, response, append(errs, fmt.Errorf("received HTTP status code %d in attempt %d", statusCode, attempt)))
		}
	}

	return response, errs
}

func (r *Request) wait(duration time.Duration) {
	if duration == 0 {
		return
	}

	ctx, ctxFunc := context.WithTimeout(r.Context(), duration)
	defer ctxFunc()
	<-ctx.Done()
}

// WithRetryPolicy sets the retry policy for the request.
func WithRetryPolicy(retries int, duration time.Duration, policy FallbackPolicy, statuscodes ...int) RequestOption {
	return func(request *Request) (err error) {
		if retries < 0 {
			retries = 0
		} else if retries > 10 {
			retries = 10
		}

		request.Retries = retries
		request.FallbackDuration = duration
		request.FallbackPolicy = policy
		request.FallbackStatusCodes = statuscodes

		return nil
	}
}

// WithTimeout sets the timeout duration for the request.
func WithTimeout(duration time.Duration) RequestOption {
	return func(request *Request) (err error) {
		request.Timeout = duration
		return nil
	}
}

// WithRequestOptions composes multiple request options.
func WithRequestOptions(opts ...RequestOption) RequestOption {
	return func(request *Request) (err error) {
		for _, opt := range opts {
			err = errors.Join(err, opt(request))
		}

		return err
	}
}

// WithURL sets the URL for the request.
func WithURL(rawUrl string) RequestOption {
	return func(request *Request) (err error) {
		parsedUrl, err := url.Parse(rawUrl)
		if err != nil {
			return err
		}

		request.URL = parsedUrl
		return nil
	}
}

// WithURLQuery sets the URL query parameters for the request.
func WithURLQuery(query map[string][]any) RequestOption {
	return func(request *Request) error {
		url := request.URL.Query()
		for key, values := range query {
			for _, value := range values {
				url.Add(key, fmt.Sprint(value))
			}
		}

		request.URL.RawQuery = url.Encode()
		return nil
	}
}

// WithBody sets the request body.
func WithBody(body io.Reader) RequestOption {
	return func(request *Request) error {
		buffer := &bytes.Buffer{}
		size, err := io.Copy(buffer, body)
		if err != nil {
			return err
		}

		request.Body = io.NopCloser(buffer)
		request.ContentLength = size
		return nil
	}
}

// WithBodyXML XML serializes the object and sets the request body as XML.
func WithBodyXML(object any) RequestOption {
	return func(request *Request) error {
		body, err := xml.MarshalIndent(object, "", "  ")
		if err != nil {
			return err
		}

		if err = WithBody(bytes.NewReader(body))(request); err != nil {
			return err
		}

		request.Header.Add("Content-Type", "application/xml")
		return nil
	}
}

// WithBodyJSON JSON serializes the object and sets the request body as JSON.
func WithBodyJSON(object any) RequestOption {
	return func(request *Request) error {
		body, err := json.Marshal(object)
		if err != nil {
			return err
		}

		if err = WithBody(bytes.NewReader(body))(request); err != nil {
			return err
		}

		request.Header.Add("Content-Type", "application/json")
		return nil
	}
}

// WithBodyFormURLEncoded sets the request body as form-urlencoded.
func WithBodyFormURLEncoded(form map[string][]string) RequestOption {
	return func(request *Request) error {
		formValues := url.Values{}
		for key, values := range form {
			for _, value := range values {
				formValues.Add(key, value)
			}
		}

		if err := WithBody(strings.NewReader(formValues.Encode()))(request); err != nil {
			return err
		}

		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		return nil
	}
}

// WithBodyFormData writes the content to body using the multipart
// writer.
func WithBodyFormData(form map[string][]byte) RequestOption {
	return func(request *Request) error {
		body := bytes.Buffer{}
		mWriter := multipart.NewWriter(&body)
		for key, value := range form {
			writer, err := mWriter.CreateFormField(key)
			if err != nil {
				return err
			}

			if _, err = writer.Write(value); err != nil {
				return err
			}
		}

		mWriter.Close()
		if err := WithBody(&body)(request); err != nil {
			return err
		}

		request.Header.Add("Content-Type", mWriter.FormDataContentType())
		return nil
	}
}

// WithBodyFormDataFile reads the given files and writes it as multipart form.
// the functional options allows you to mutate the file content before it's being written.
func WithBodyFormDataFile(filePath, field string, opts ...func(content []byte) []byte) RequestOption {
	return func(request *Request) (err error) {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		for _, opt := range opts {
			content = opt(content)
		}

		return WithBodyFormData(map[string][]byte{
			field: content,
		})(request)
	}
}

// WithAuthorizationBasic encodes the credentials with basic HTTP authentication.
// It sets the valkue in the Authorization HTTP header.
func WithAuthorizationBasic(username, password string) RequestOption {
	return func(request *Request) error {
		auth := fmt.Sprintf("%s:%s", username, password)
		cred := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
		request.Header.Add("Authorization", cred)
		return nil
	}
}

// WithAuthorizationBearer executes the callback to fetch a token, the token from
// the result will be set in the Authorization header
func WithAuthorizationBearer(fn func(ctx context.Context) (string, error)) RequestOption {
	return func(request *Request) error {
		token, err := fn(request.Context())
		if err != nil {
			return err
		}

		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
		return nil
	}
}

// WithHeader sets key value as HTTP header in the request.
func WithHeader(key string, value any) RequestOption {
	return func(request *Request) error {
		request.Header.Add(key, fmt.Sprint(value))
		return nil
	}
}
