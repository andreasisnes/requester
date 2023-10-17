![coverage](https://raw.githubusercontent.com/andreasisnes/requester/badges/.badges/coverage.svg)

# Requester
Reqester is a Go package that provides a wrapper around the standard HTTP package, utilizing the function options pattern for handling requests and responses. It offers additional features for handling HTTP requests, including retries, fallback policies, and more.

In the realm of HTTP client libraries for Go, developers often find themselves working with the standard HTTP client, request, and response pattern. While effective, this pattern can become verbose, especially when dealing with error handling and mutating requests.

Many projects and libraries in the Go ecosystem follow the builder pattern, which can provide a cleaner and more fluent API for constructing requests. However, the builder pattern can sometimes be challenging to extend with additional functionality.

In contrast, the functional pattern offers a flexible and composable approach to working with HTTP requests. This pattern aligns well with Go's idiomatic style, allowing developers to easily apply modifications and customizations to requests through functional options.

## Installation
```bash
go get github.com/andreasisnes/requester
```

## Usage

```go
import (
    "github.com/andreasisnes/requester"
    "net/http"
) 

// Create a new client with default options
client := requester.New()

// Modify client options
clientWithOptions := requester.New(
    requester.WithBaseURL("https://api.example.com"),
    requester.WithClient(&http.Client{}),
)
```