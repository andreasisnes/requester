package rejester

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ResponseOption func(request *Response) error

type Response struct {
	*http.Response
	Err error
}

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

func WithStatusCodeAssertion(statusCodes ...int) ResponseOption {
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

func WithUnmarshalJSON[T any](object *T, statuscodes ...int) ResponseOption {
	return func(response *Response) error {
		return WithBodyUnmarshal(object, json.Unmarshal, statuscodes...)(response)
	}
}

func WithUnmarshalXML[T any](object *T, statuscodes ...int) ResponseOption {
	return func(response *Response) error {
		return WithBodyUnmarshal(object, xml.Unmarshal, statuscodes...)(response)
	}
}

func WithBodyUnmarshal[T any](object *T, unmarshaler func(data []byte, v any) error, statuscodes ...int) ResponseOption {
	return func(response *Response) (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf(fmt.Sprint(p))
			}
		}()

		for _, code := range statuscodes {
			if response.StatusCode == code {
				if object == nil {
					object = new(T)
				}

				if response.Body != nil {
					body, err := io.ReadAll(response.Body)
					if err != nil {
						return err
					}

					response.Body = io.NopCloser(bytes.NewBuffer(body))
					return unmarshaler(body, object)
				}
			}
		}

		return err
	}
}
