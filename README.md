<div align="center">

[![Pipeline](https://github.com/andreasisnes/requester/actions/workflows/pipeline.yml/badge.svg)](https://github.com/andreasisnes/requester/actions/workflows/pipeline.yml)
![coverage](https://raw.githubusercontent.com/andreasisnes/requester/badges/.badges/coverage.svg)
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
This package extends the functionality of the standard http package by incorporating and enhancing its core components: the Client, Request, and Response objects. It introduces additional fields, methods, and object compositions, augmenting the capabilities provided by the standard HTTP package.

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
	sdk := &EchoServer{
		client: requester.New(
			requester.WithBaseURL("https://postman-echo.com"),
		),
	}

	ctx := context.Background()

	response, _ := sdk.GET(ctx, map[string][]any{
		"timstamp": {time.Now()},
	})
	PrintJSON(response)

	response, _ = sdk.POST(ctx, map[string]any{
		"test": 123,
	})
	PrintJSON(response)
}

func PrintJSON(response Response) {
	jsonBytes, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println(string(jsonBytes))
}

func (sdk *EchoServer) POST(ctx context.Context, payload map[string]any) (out Response, err error) {
	return out, sdk.client.POST(ctx, "post").
		Do(
			requester.WithRequestJSON(payload),
			requester.WithRequestHeader("X-Custom", 123),
			requester.WithRequestAuthorizationBasic("admin", "password"),
		).
		Handle(
			requester.WithResponseJSON(&out),
		)
}

func (sdk *EchoServer) GET(ctx context.Context, query map[string][]any) (out Response, err error) {
	return out, sdk.client.GET(ctx, "get").
		Do(
			requester.WithRequestURLQuery(query),
		).
		Handle(
			requester.WithResponseJSON(&out),
		)
}
```


The first request GET will write following to stdout.
```json
{
  "args": {
    "timstamp": "2023-10-19 13:05:34.994444724 +0200 CEST m=+0.000163639"
  },
  "data": null,
  "headers": {
    "accept-encoding": "gzip",
    "host": "postman-echo.com",
    "user-agent": "Go-http-client/2.0",
    "x-amzn-trace-id": "Root=1-65310d53-0db6144a5175aea0772afad7",
    "x-forwarded-port": "443",
    "x-forwarded-proto": "https"
  },
  "url": "https://postman-echo.com/get?timstamp=2023-10-19+13%3A05%3A34.994444724+%2B0200+CEST+m%3D%2B0.000163639"
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
    "authorization": "Basic YWRtaW46cGFzc3dvcmQ=",
    "content-length": "12",
    "content-type": "application/json",
    "host": "postman-echo.com",
    "user-agent": "Go-http-client/2.0",
    "x-amzn-trace-id": "Root=1-65310d53-14c9787b400c18fc69ec0f40",
    "x-custom": "123",
    "x-forwarded-port": "443",
    "x-forwarded-proto": "https"
  },
  "url": "https://postman-echo.com/post"
}
```

# Contributing
If you want to contribute, you are welcome to open an [issue](https://github.com/andreasisnes/requester/issues) or submit a [pull request](https://github.com/andreasisnes/requester/pulls).

