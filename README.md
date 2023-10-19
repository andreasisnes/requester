<div align="center">

[![Pipeline](https://github.com/andreasisnes/requester/actions/workflows/pipeline.yml/badge.svg)](https://github.com/andreasisnes/requester/actions/workflows/pipeline.yml)
![coverage](https://raw.githubusercontent.com/andreasisnes/requester/badges/.badges/main/coverage.svg)
![GitHub](https://img.shields.io/github/license/andreasisnes/requester)
[![Go Report Card](https://goreportcard.com/badge/github.com/andreasisnes/requester)](https://goreportcard.com/report/github.com/andreasisnes/requester)
[![GoDoc](https://godoc.org/github.com/andreasisnes/requester?status.svg)](https://godoc.org/github.com/andreasisnes/requester)

</div>

# Requester
Requester is a Go package that provides a wrapper around the standard HTTP package, utilizing the function options pattern for handling requests and responses. It offers additional features for handling HTTP requests, including retries, fallback policies, and more.

In the realm of HTTP client libraries for Go, developers often find themselves working with the standard HTTP client, request, and response pattern. While effective, this pattern can become verbose, especially when dealing with error handling and mutating requests.

Many projects and libraries in the Go ecosystem follow the builder pattern, which can provide a cleaner and more fluent API for constructing requests. However, the builder pattern can sometimes be challenging to extend with additional functionality.

In contrast, the functional pattern offers a flexible and composable approach to working with HTTP requests. This pattern aligns well with Go's idiomatic style, allowing developers to easily apply modifications and customizations to requests through functional options.

## Components
This package extends the functionality of the standard http package by incorporating and enhancing its core components: the Client, Request, and Response objects. It introduces additional fields, methods, and object compositions of the standard http package, augmenting the capabilities provided by the standard HTTP package with some extra stuff.

Errors are propagated through the client initialization, request construction, sending, and handling. If any errors occur along this process, subsequent steps will not execute the provided callbacks. Therefore you omit checking the errors for each stage. See the usage section below.


## Installation
```bash
go get github.com/andreasisnes/requester
```

## Usage

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/andreasisnes/requester"
)

type Response struct {
	Args    map[string]string `json:"args"`
	Data    map[string]any    `json:"data"`
	Headers map[string]string `json:"headers"`
	URL     string            `json:"url"`
}

type EchoServer struct {
	client *requester.Client
}

func main() {
	echo := &EchoServer{
		client: requester.New(
			requester.WithBaseURL("https://postman-echo.com"),
		),
	}

	ctx := context.Background()

	response, _ := echo.GET(ctx, map[string][]any{
		"timstamp": {time.Now()},
	})
	PrintJSON(response)

	response, _ = echo.POST(ctx, map[string]any{
		"test": 123,
	})
	PrintJSON(response)
}

func PrintJSON(response Response) {
	jsonBytes, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println(string(jsonBytes))
}

func (echo *EchoServer) GET(ctx context.Context, query map[string][]any) (out Response, err error) {
	return out, echo.client.GET(ctx, "get").
		Do(
			echo.WithDefaultRequestOptions,
			requester.WithRequestURLQuery(query),
			func(request *requester.Request) (err error) {
				request.Header.Add("X-Header", "1337")
				return nil
			},
		).
		Handle(
			requester.WithResponseStatusCodeAssertion(http.StatusOK),
			requester.WithResponseJSON(&out, http.StatusOK),
		)
}

func (echo *EchoServer) POST(ctx context.Context, payload map[string]any) (out Response, err error) {
	return out, echo.client.POST(ctx, "post").
		Do(
			echo.WithDefaultRequestOptions,
			requester.WithRequestJSON(payload),
		).
		Handle(
			requester.WithResponseStatusCodeAssertion(http.StatusOK),
			requester.WithResponseJSON(&out, http.StatusOK),
		)
}

func (echo *EchoServer) WithDefaultRequestOptions(request *requester.Request) error {
	return requester.WithRequestOptions(
		requester.WithRequestAuthorizationBasic("username", "password"),
		requester.WithRequestTimeout(time.Second),
		requester.WithRequestRetryPolicy(3, time.Second, requester.FallbackPolicyExponential),
	)(request)
}
```


The first request GET will write following to stdout.
```json
{
  "args": {
    "timstamp": "2023-10-19 17:03:50.688733922 +0200 CEST m=+0.000166782"
  },
  "data": null,
  "headers": {
    "accept-encoding": "gzip",
    "authorization": "Basic dXNlcm5hbWU6cGFzc3dvcmQ=",
    "host": "postman-echo.com",
    "user-agent": "Go-http-client/2.0",
    "x-amzn-trace-id": "Root=1-6531452b-640a2bf95549969230ae35d5",
    "x-forwarded-port": "443",
    "x-forwarded-proto": "https",
    "x-header": "1337"
  },
  "url": "https://postman-echo.com/get?timstamp=2023-10-19+17%3A03%3A50.688733922+%2B0200+CEST+m%3D%2B0.000166782"
}
```

The second request POST will write the following to stdout.
```json
{
  "args": {},
  "data": {
    "test": 123
  },
  "headers": {
    "accept-encoding": "gzip",
    "authorization": "Basic dXNlcm5hbWU6cGFzc3dvcmQ=",
    "content-length": "12",
    "content-type": "application/json",
    "host": "postman-echo.com",
    "user-agent": "Go-http-client/2.0",
    "x-amzn-trace-id": "Root=1-6531452b-0d33cc866fa05f4758a9fbfe",
    "x-forwarded-port": "443",
    "x-forwarded-proto": "https"
  },
  "url": "https://postman-echo.com/post"
}
```

# Contributing
If you want to contribute, you are welcome to open an [issue](https://github.com/andreasisnes/requester/issues) or submit a [pull request](https://github.com/andreasisnes/requester/pulls).

