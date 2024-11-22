package multiproxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"
)

type ReverseProxyOptions[T any] struct {
	lb *LoadBalancer[T]
}

func newReverseProxy[T any](opts ReverseProxyOptions[T], upstream *Upstream[T]) *httputil.ReverseProxy {
	// no upstreams available
	if upstream == nil {
		fmt.Printf("No server available to handle this request. this should not happen")
		return nil
	}

	now := time.Now()

	errHand := func(proxy *httputil.ReverseProxy, opts ReverseProxyOptions[T]) func(http.ResponseWriter, *http.Request, error) {
		return func(w http.ResponseWriter, r *http.Request, err error) {

			if opts.lb.OnError != nil {
				opts.lb.OnError(upstream, r, err)
			}

			opts.lb.ReportError(upstream, err, r)

			healthy := IsHealthyResponse(r, err)
			upstream.Healthy = healthy

			if !healthy {
				if opts.lb.OnMarkUnhealthy != nil {
					opts.lb.OnMarkUnhealthy(upstream)
				}
			}

			// upstream was marked as unhealthy, try the next server
			if !upstream.Healthy {
				nextUpstream := opts.lb.Random(r)
				success := false
				// try a few times to get a different server
				for i := 0; i < 20; i++ {
					if nextUpstream == nil || nextUpstream.Equal(upstream) {
						// we ended up getting the one that errored, try again
						nextUpstream = opts.lb.Random(r)
					} else {
						success = true
						break
					}
				}

				// could not find a different server than the one that just errored
				if !success || nextUpstream == nil {
					http.Error(w, "No server available", http.StatusServiceUnavailable)
					return
				}

				// try the nextUpstream server
				handler := newReverseProxy(opts, nextUpstream)
				if handler != nil {
					handler.ServeHTTP(w, r)
				} else {
					http.Error(w, "No server available", http.StatusServiceUnavailable)
				}
			}
		}
	}

	proxy := &httputil.ReverseProxy{
		ModifyResponse: func(response *http.Response) error {

			if opts.lb.AfterRequest != nil {
				opts.lb.AfterRequest(upstream, response.Request, response)
			}

			upstream.OnFinishRequest(now)
			return nil
		},
		Director: func(req *http.Request) {
			if upstream == nil {
				return
			}

			upstream.OnBeginRequest()

			now = time.Now()
			targetQuery := upstream.Url.RawQuery
			req.URL.Scheme = upstream.Url.Scheme
			req.URL.Host = upstream.Url.Host
			req.URL.Path, req.URL.RawPath = joinURLPath(upstream.Url, req.URL)
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}

			if opts.lb.BeforeRequest != nil {
				opts.lb.BeforeRequest(upstream, req)
			}
		},
	}

	proxy.ErrorHandler = errHand(proxy, opts)

	return proxy
}

func NewReverseProxyHandler[T any](lb *LoadBalancer[T]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		upstream := lb.Random(r)
		if upstream == nil {
			http.Error(w, "No server available to handle this request.", http.StatusServiceUnavailable)
			return
		}
		p := newReverseProxy(ReverseProxyOptions[T]{lb: lb}, upstream)
		p.ServeHTTP(w, r)
	}
}
