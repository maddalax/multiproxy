## Multiproxy - Go Reverse Proxy with Load Balancing

**Multiproxy** is a flexible, open-source reverse proxy written in Go that enables load balancing across multiple upstreams. It offers host and path-based routing, automatic health checks, and failure recovery.

## Features

- **Load Balancing**: Distributes requests across healthy upstream servers using a randomized strategy.
- **Host and Path Routing**: Route requests based on specified hostnames and paths.
- **Automatic Health Checks**: Continuously monitors upstream servers and automatically handles failures.
- **Customizable Hooks**: Customize request handling with hooks for before and after requests, error handling, and upstream health status changes.

## Installation

To use Multiproxy in your Go project, get the package:

``
go get github.com/maddalax/multiproxy
``

## Key Concepts

### Load Balancer

- **CreateLoadBalancer**: Initializes a new load balancer that manages upstream servers.
- **Add**: Adds an upstream server to the load balancer.
- **Hooks**:
  - **BeforeRequest**: A function called before sending a request to an upstream.
  - **AfterRequest**: A function called after receiving a response from an upstream.
  - **OnError**: A function called when a request fails to an upstream.
  - **OnMarkHealthy**: A function called when an upstream is marked as healthy.
  - **OnMarkUnhealthy**: A function called when an upstream is marked as unhealthy.

### Upstream Health Management

- Upstreams are automatically marked as **unhealthy** for 10 seconds if a request fails with specific error statuses, such as:
  - `StatusBadGateway`
  - `StatusGatewayTimeout`
  - `StatusServiceUnavailable`
  - Connection refused
- After 10 seconds, unhealthy upstreams are tested to see if they have recovered.

### Routing Rules

- **Host and Path Matching**: Route requests to upstreams based on specified host and path criteria.
- Upstreams can be configured with `Matches` to determine which requests they can service.

Example Usage:
```go
package main

import (
	"fmt"
	"multiproxy"
	"net/http"
	"net/url"
	"time"
)

func main() {
	// Create a new load balancer
	lb := multiproxy.CreateLoadBalancer()

	// Setup custom hooks
	lb.BeforeRequest = func(up *multiproxy.Upstream, req *http.Request) {
		req.Header.Add("X-Forwarded-For", "")
	}

	lb.AfterRequest = func(up *multiproxy.Upstream, req *http.Request, res *http.Response) {
	}

	lb.OnError = func(up *multiproxy.Upstream, req *http.Request, err error) {
	}

	lb.OnMarkUnhealthy = func(up *multiproxy.Upstream) {
		//slog.Info("Upstream marked as unhealthy", slog.String("host", up.Url.Host))
	}

	lb.OnMarkHealthy = func(up *multiproxy.Upstream) {
		//slog.Info("Upstream marked as healthy", slog.String("host", up.Url.Host))
	}

	// Add upstreams to the load balancer
	upstreamUrls := []string{
		"http://localhost:4001",
		"http://localhost:4000",
		"http://localhost:4003",
	}

	for _, urlStr := range upstreamUrls {
		upstreamUrl, _ := url.Parse(urlStr)
		lb.Add(&multiproxy.Upstream{
			Url:     upstreamUrl,
			Healthy: true,
			Matches: []multiproxy.Match{
				{Host: "paas.example.com", Path: ""},
			},
		})
	}

	// Start the reverse proxy
	handler := multiproxy.NewReverseProxyHandler(lb)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":80", nil)
}
```

License

This project is licensed under the MIT License.
