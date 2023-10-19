package requester

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ResponseOption is a callback signature for modifying response options.
type ResponseOption func(request *Response) error

// Response is a wrapper around the standard http.Response providing additional features.
type Response struct {
	*http.Response
	Err error
}

// Handle executes the response handling options.
// If there is an error associated with the response, it returns that error.
func (r *Response) Handle(opts ...ResponseOption) error {
	if r.Err != nil {
		return r.Err
	}

	var err error
	for _, o := range opts {
		err = errors.Join(r.Err, o(r))
	}

	return err
}

// WithResponseStatusCodeAssertion checks if the response status code matches any of the specified codes.
// If it does, it returns nil. Otherwise, it provides an error message.
func WithResponseStatusCodeAssertion(statusCodes ...int) ResponseOption {
	return func(response *Response) error {
		for _, code := range statusCodes {
			if code == response.StatusCode {
				return nil
			}
		}

		if response.Body != nil {
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return err
			}

			response.Body = io.NopCloser(bytes.NewBuffer(body))
			if len(body) > 0 {
				return fmt.Errorf(string(body))
			}
		}

		return fmt.Errorf("expected status code(s) '%v', received '%d'", statusCodes, response.StatusCode)
	}
}

// WithResponseJSON unmarshals the JSON response body to an object.
// The object parameter should be a pointer to the target type. It will
// only attempt to deserialize the payload if the response has one of the provided status codes.
// If the list of status codes is empty, it will attempt to deserialize for all status codes.
func WithResponseJSON[T any](object *T, statuscodes ...int) ResponseOption {
	return func(response *Response) error {
		return WithResponseBody(object, json.Unmarshal, statuscodes...)(response)
	}
}

// WithResponseXML unmarshals the XML response body to an object.
// The object parameter should be a pointer to the target type. It will
// only attempt to deserialize the payload if the response has one of the provided status codes.
// If the list of status codes is empty, it will attempt to deserialize for all status codes.
func WithResponseXML[T any](object *T, statuscodes ...int) ResponseOption {
	return func(response *Response) error {
		return WithResponseBody(object, xml.Unmarshal, statuscodes...)(response)
	}
}

// WithUnmarshalXML unmarshals the response body to an object using the given unmarshaler.
// The object parameter should be a pointer to the target type. It will
// only attempt to deserialize the payload if the response has one of the provided status codes.
// If the list of status codes is empty, it will attempt to deserialize for all status codes.
func WithResponseBody[T any](object *T, unmarshaler func(data []byte, v any) error, statuscodes ...int) ResponseOption {
	return func(response *Response) (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf(fmt.Sprint(p))
			}
		}()

		deserialize := func() error {
			if response.Body != nil {
				body, err := io.ReadAll(response.Body)
				if err != nil {
					return err
				}

				response.Body = io.NopCloser(bytes.NewBuffer(body))
				return unmarshaler(body, object)
			}

			return nil
		}

		if len(statuscodes) == 0 {
			return deserialize()
		}

		for _, code := range statuscodes {
			if response.StatusCode == code {
				return deserialize()
			}
		}

		return nil
	}
}
