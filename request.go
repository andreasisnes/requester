package rejester

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

type FallbackPolicy int

const (
	FallbackPolicyLinear FallbackPolicy = iota
	FallbackPolicyExponential
)

type RequestOption func(request *Request) (err error)

type Request struct {
	*http.Request
	*http.Client
	Err                 error
	Retries             int
	FallbackPolicy      FallbackPolicy
	FallbackStatusCodes []int
}

func (r *Request) Dry(opts ...RequestOption) (err error) {
	if r.Err != nil {
		return r.Err
	}

	for _, o := range opts {
		err = errors.Join(r.Err, o(r))
	}

	return err
}

func (r *Request) Do(opts ...RequestOption) *Response {
	if r.Err != nil || r.Request == nil {
		return &Response{Response: &http.Response{}, Err: r.Err}
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
			r.wait(time.Second * (time.Duration(attempt * attempt)))
		default:
			r.wait(time.Second * time.Duration(attempt))
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
	ctx, ctxFunc := context.WithTimeout(r.Context(), duration)
	defer ctxFunc()
	<-ctx.Done()
}

func WithRetryPolicy(retries int, policy FallbackPolicy, statuscodes ...int) RequestOption {
	return func(request *Request) (err error) {
		if retries < 0 {
			retries = 0
		} else if retries > 10 {
			retries = 10
		}

		request.Retries = retries
		request.FallbackPolicy = policy
		request.FallbackStatusCodes = statuscodes
		return nil
	}
}

func WithTimeout(duration time.Duration) RequestOption {
	return func(request *Request) (err error) {
		request.Timeout = duration
		return nil
	}
}

func WithRequestOptions(opts ...RequestOption) RequestOption {
	return func(request *Request) (err error) {
		for _, opt := range opts {
			err = errors.Join(err, opt(request))
		}

		return err
	}
}

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

func WithAuthorizationBasic(username, password string) RequestOption {
	return func(request *Request) error {
		auth := fmt.Sprintf("%s:%s", username, password)
		cred := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
		request.Header.Add("Authorization", cred)
		return nil
	}
}

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

func WithHeader(key string, value any) RequestOption {
	return func(request *Request) error {
		request.Header.Add(key, fmt.Sprint(value))
		return nil
	}
}
