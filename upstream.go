package multiproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type Match struct {
	// If host is empty, it matches all hosts, and will only match if the path matches
	Host string
	// If path is empty, it matches all paths, and will only match if the host matches
	Path string
}

type Upstream struct {
	Url                 *url.URL
	LastRequest         time.Time
	AverageResponseTime time.Duration
	TotalRequests       int
	TotalErrorResponses int
	Healthy             bool
	// MatchesFunc is a custom function that can be used to determine if a request matches an upstream
	MatchesFunc func(req *http.Request, match *Match) bool
	Matches     []Match
}

func (u *Upstream) Handle(proxy *httputil.ReverseProxy, w http.ResponseWriter, r *http.Request) {
	proxy.ServeHTTP(w, r)
}

// CanServiceRequest returns true if the upstream has any matches for the given URL
func (u *Upstream) CanServiceRequest(req *http.Request) bool {

	if !u.Healthy {
		return false
	}

	if u.MatchesReq(req) {
		return true
	}

	return false
}

func (u *Upstream) Equal(u2 *Upstream) bool {
	return u.Url.Host == u2.Url.Host && u.Url.Scheme == u2.Url.Scheme
}
