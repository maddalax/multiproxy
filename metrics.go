package multiproxy

import (
	"log/slog"
	"net/http"
	"time"
)

func (lb *LoadBalancer[T]) PrintMetrics() {
	for i, upstream := range lb.upstreams {
		slog.Info("Upstream metrics",
			slog.Int("index", i),
			slog.String("host", upstream.Url.Host),
			slog.Bool("healthy", upstream.Healthy),
			slog.Duration("average_response_time", upstream.AverageResponseTime),
			slog.Int64("total_requests", upstream.TotalRequests.Load()),
			slog.Int("total_error_responses", upstream.TotalErrorResponses))
	}
}

func (lb *LoadBalancer[T]) ReportError(upstream *Upstream[T], err error, r *http.Request) {
	upstream.TotalErrorResponses++
	if r.Response == nil {
		return
	}
}

func (u *Upstream[T]) OnBeginRequest() {
	u.LastRequest = time.Now()
	u.TotalRequests.Add(1)
}

func (u *Upstream[T]) OnFinishRequest(start time.Time) {
	took := time.Since(start)
	alpha := 0.1
	u.AverageResponseTime = time.Duration(
		alpha*float64(took) + (1-alpha)*float64(u.AverageResponseTime),
	)
}
