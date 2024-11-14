package multiproxy

import (
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type LoadBalancer struct {
	upstreams       []*Upstream
	OnError         func(up *Upstream, req *http.Request, err error)
	BeforeRequest   func(up *Upstream, req *http.Request)
	AfterRequest    func(up *Upstream, req *http.Request, res *http.Response)
	OnMarkUnhealthy func(up *Upstream)
	OnMarkHealthy   func(up *Upstream)
}

func CreateLoadBalancer() *LoadBalancer {
	lb := &LoadBalancer{}
	go lb.StartHealthWatcher()
	return lb
}

func (lb *LoadBalancer) StartHealthWatcher() {
	healthyTicker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-healthyTicker.C:
			for _, upstream := range lb.upstreams {
				// if the upstream is unhealthy, and it's been more than 10 seconds since the last request
				// mark it as healthy to see if it's back up
				// it will be marked as unhealthy again if the next request fails
				if !upstream.Healthy && time.Since(upstream.LastRequest) > 10*time.Second {
					upstream.Healthy = true
					if lb.OnMarkHealthy != nil {
						lb.OnMarkHealthy(upstream)
					}
				}
			}
		}
	}
}

func (lb *LoadBalancer) Add(upstream *Upstream) {
	upstream.Healthy = true
	lb.upstreams = append(lb.upstreams, upstream)
}

// GetValidUpstreams returns a list of upstreams that are healthy and can service the incoming request
func (lb *LoadBalancer) GetValidUpstreams(req *http.Request) []*Upstream {
	var upstreams = make([]*Upstream, 0, len(lb.upstreams))

	for _, upstream := range lb.upstreams {
		if upstream.CanServiceRequest(req) {
			upstreams = append(upstreams, upstream)
		}
	}

	return upstreams
}

func (lb *LoadBalancer) GetUpstreamMatching(url *url.URL) *Upstream {
	for _, upstream := range lb.upstreams {
		if upstream.Url.Host == url.Host && upstream.Url.Scheme == url.Scheme {
			return upstream
		}
	}
	return nil
}

func (lb *LoadBalancer) Random(r *http.Request) *Upstream {
	upstreams := lb.GetValidUpstreams(r)
	l := len(upstreams)
	if l == 0 {
		return nil
	}
	index := rand.Intn(l)
	return upstreams[index]
}
