package multiproxy

import (
	"log/slog"
	"net/http"
	"time"
)

func (lb *LoadBalancer) PrintMetrics() {
	for i, upstream := range lb.upstreams {
		slog.Info("Upstream metrics",
			slog.Int("index", i),
			slog.String("host", upstream.Url.Host),
			slog.Bool("healthy", upstream.Healthy),
			slog.Duration("average_response_time", upstream.AverageResponseTime),
			slog.Int("total_requests", upstream.TotalRequests),
			slog.Int("total_error_responses", upstream.TotalErrorResponses))
	}
}

func (lb *LoadBalancer) ReportError(upstream *Upstream, err error, r *http.Request) {
	upstream.TotalErrorResponses++
	if r.Response == nil {
		return
	}
}

func (u *Upstream) OnBeginRequest() {
	u.LastRequest = time.Now()
	u.TotalRequests++
}

func (u *Upstream) OnFinishRequest(start time.Time) {
	took := time.Since(start)
	alpha := 0.1
	u.TotalRequests++
	u.AverageResponseTime = time.Duration(
		alpha*float64(took) + (1-alpha)*float64(u.AverageResponseTime),
	)
}
