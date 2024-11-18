package multiproxy

import (
	"net/http"
	"strings"
)

func (u *Upstream) PathMatches(path string, match *Match) bool {
	if match.Path == "" {
		return true
	}
	return strings.HasPrefix(path, match.Path)
}

func (u *Upstream) HostMatches(host string, match *Match) bool {
	if match.Host == "" {
		return true
	}
	return host == match.Host
}

func (u *Upstream) MatchesReq(req *http.Request) bool {
	for _, match := range u.Matches {
		if u.MatchesFunc != nil {
			if u.MatchesFunc(req, &match) {
				return true
			}
		} else {
			if u.PathMatches(req.URL.Path, &match) && u.HostMatches(req.Host, &match) {
				return true
			}
		}
	}
	return false
}
