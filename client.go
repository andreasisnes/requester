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

func New(opts ...ClientOptions) *Client {
	c := &Client{
		Client: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithClient(httpClient *http.Client) ClientOptions {
	return func(client *Client) {
		client.Client = httpClient
	}
}

func WithBaseURL(url string) ClientOptions {
	return func(client *Client) {
		client.url = url
	}
}

func (c *Client) DELETE(ctx context.Context, route ...string) *Request {
	return c.Request(ctx, http.MethodDelete, route...)
}

func (c *Client) PUT(ctx context.Context, route ...string) *Request {
	return c.Request(ctx, http.MethodPut, route...)
}

func (c *Client) GET(ctx context.Context, route ...string) *Request {
	return c.Request(ctx, http.MethodGet, route...)
}

func (c *Client) POST(ctx context.Context, route ...string) *Request {
	return c.Request(ctx, http.MethodPost, route...)
}

func (c *Client) PATCH(ctx context.Context, route ...string) *Request {
	return c.Request(ctx, http.MethodPatch, route...)
}

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
