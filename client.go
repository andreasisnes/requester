package rejester

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	*http.Client
	url string
}

type ClientOptions func(client *Client)

// New initialize a default client.
// Provide ClientOptions to modify default behavior
func New(opts ...ClientOptions) *Client {
	c := &Client{
		Client: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithClient sets client to given HTTP client instance
func WithClient(httpClient *http.Client) ClientOptions {
	return func(client *Client) {
		client.Client = httpClient
	}
}

// WithBaseURL sets a base URL which will be the prefix for
// all outbound HTTP requests
func WithBaseURL(url string) ClientOptions {
	return func(client *Client) {
		client.url = url
	}
}

// DELETE creates a HTTP DELETE request with the given route. If a base URL is specified
// in client the given route should just contain the path otherwise the whole URL. The route
// segments will be joined with / as seperator.
func (c *Client) DELETE(ctx context.Context, route ...string) *Request {
	return c.Request(ctx, http.MethodDelete, route...)
}

// PUT creates a HTTP PUT request with the given route. If a base URL is specified
// in client the given route should just contain the path otherwise the whole URL. The route
// segments will be joined with / as seperator.
func (c *Client) PUT(ctx context.Context, route ...string) *Request {
	return c.Request(ctx, http.MethodPut, route...)
}

// GET creates a HTTP GET request with the given route. If a base URL is specified
// in client the given route should just contain the path otherwise the whole URL. The route
// segments will be joined with / as seperator.
func (c *Client) GET(ctx context.Context, route ...string) *Request {
	return c.Request(ctx, http.MethodGet, route...)
}

// POST creates a HTTP POST request with the given route. If a base URL is specified
// in client the given route should just contain the path otherwise the whole URL. The route
// segments will be joined with / as seperator.
func (c *Client) POST(ctx context.Context, route ...string) *Request {
	return c.Request(ctx, http.MethodPost, route...)
}

// PATCH creates a HTTP PATCH request with the given route. If a base URL is specified
// in client the given route should just contain the path otherwise the whole URL. The route
// segments will be joined with / as seperator.
func (c *Client) PATCH(ctx context.Context, route ...string) *Request {
	return c.Request(ctx, http.MethodPatch, route...)
}

// Request creates a HTTP request with the given HTTP method and route. If a base URL is specified
// in client the given route should just contain the path otherwise the whole URL. The route
// segments will be joined with / as seperator.
func (c *Client) Request(ctx context.Context, method string, routes ...string) *Request {
	uri, err := func() (string, error) {
		if c.url == "" {
			return strings.Join(routes, "/"), nil
		}

		return url.JoinPath(c.url, routes...)
	}()

	request, e := http.NewRequestWithContext(ctx, method, uri, nil)
	if e != nil {
		err = errors.Join(err, e)
	}

	return &Request{Request: request, Client: c.Client, Err: err}
}
