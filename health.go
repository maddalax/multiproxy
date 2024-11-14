package multiproxy

import (
	"errors"
	"net"
	"net/http"
)

func IsHealthyResponse(r *http.Request, err error) bool {
	if r.Response != nil {
		switch r.Response.StatusCode {
		case http.StatusBadGateway:
		case http.StatusGatewayTimeout:
		case http.StatusServiceUnavailable:
			return false
		}
	}

	var e *net.OpError
	switch {
	// failed to connect, try the next server
	case errors.As(err, &e):
		// mark the upstream as unhealthy
		return false
	}

	return true
}
