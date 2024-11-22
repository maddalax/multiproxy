package multiproxy

import (
	"context"
	"math/rand"
	"net/http"
	"time"
)

type LoadBalancer[T any] struct {
	upstreams []*Upstream[T]
	// upstreams that are being prepared to be added to the load balancer
	stagedUpstreams []*Upstream[T]
	OnError         func(up *Upstream[T], req *http.Request, err error)
	BeforeRequest   func(up *Upstream[T], req *http.Request)
	AfterRequest    func(up *Upstream[T], req *http.Request, res *http.Response)
	OnMarkUnhealthy func(up *Upstream[T])
	OnMarkHealthy   func(up *Upstream[T])
	disposed        bool
	cancel          context.CancelFunc
	context         context.Context
}

func CreateLoadBalancer[T any]() *LoadBalancer[T] {
	ctx, cancel := context.WithCancel(context.Background())
	lb := &LoadBalancer[T]{
		context: ctx,
		cancel:  cancel,
	}
	go lb.StartHealthWatcher()
	return lb
}

func (lb *LoadBalancer[T]) SetUpstreams(upstreams []*Upstream[T]) {
	for _, upstream := range upstreams {
		upstream.Healthy = true
	}
	lb.upstreams = upstreams
}

func (lb *LoadBalancer[T]) AddStagedUpstream(upstream *Upstream[T]) {
	lb.stagedUpstreams = append(lb.stagedUpstreams, upstream)
}

func (lb *LoadBalancer[T]) ClearStagedUpstreams() {
	lb.stagedUpstreams = nil
}

func (lb *LoadBalancer[T]) ApplyStagedUpstreams() {
	lb.upstreams = lb.stagedUpstreams
	lb.stagedUpstreams = nil
}

func (lb *LoadBalancer[T]) Dispose() {
	lb.cancel()
}

func (lb *LoadBalancer[T]) StartHealthWatcher() {
	healthyTicker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-lb.context.Done():
			healthyTicker.Stop()
			return
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

func (lb *LoadBalancer[T]) Add(upstream *Upstream[T]) {
	upstream.Healthy = true
	lb.upstreams = append(lb.upstreams, upstream)
}

// GetValidUpstreams returns a list of upstreams that are healthy and can service the incoming request
func (lb *LoadBalancer[T]) GetValidUpstreams(req *http.Request) []*Upstream[T] {
	var upstreams = make([]*Upstream[T], 0, len(lb.upstreams))

	for _, upstream := range lb.upstreams {
		if upstream.CanServiceRequest(req) {
			upstreams = append(upstreams, upstream)
		}
	}

	return upstreams
}

func (lb *LoadBalancer[T]) GetUpstreams() []*Upstream[T] {
	return lb.upstreams
}

func (lb *LoadBalancer[T]) GetStagedUpstreams() []*Upstream[T] {
	return lb.stagedUpstreams
}

func (lb *LoadBalancer[T]) Random(r *http.Request) *Upstream[T] {
	upstreams := lb.GetValidUpstreams(r)
	l := len(upstreams)
	if l == 0 {
		return nil
	}
	index := rand.Intn(l)
	return upstreams[index]
}
