package multiproxy

import (
	"net/http"
	"strings"
)

func (u *Upstream[T]) PathMatches(path string, match *Match) bool {
	if match.Path == "" {
		return true
	}
	return strings.HasPrefix(path, match.Path)
}

func (u *Upstream[T]) HostMatches(host string, match *Match) bool {
	if match.Host == "" {
		return true
	}
	return host == match.Host
}

func (u *Upstream[T]) MatchesReq(req *http.Request) bool {
	if u.MatchesFunc != nil {
		return u.MatchesFunc(u, req)
	}
	for _, match := range u.Matches {
		if u.PathMatches(req.URL.Path, &match) && u.HostMatches(req.Host, &match) {
			return true
		}
	}
	return false
}
