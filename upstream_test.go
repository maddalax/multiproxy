package multiproxy

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func canService(host string, path string, upstream *Upstream) bool {
	req := http.Request{
		URL: &url.URL{
			Host: host,
			Path: path,
		},
		Host: host,
	}
	return upstream.CanServiceRequest(&req)
}

func TestCanServiceRequest_HealthyHostMatch(t *testing.T) {
	m := Match{
		Host: "example.com",
		Path: "",
	}
	upstream := &Upstream{
		Matches: []Match{m},
		Healthy: true,
	}
	assert.True(t, canService("example.com", "/", upstream))
}

func TestCanServiceRequest_HealthyHostAndPathMatch(t *testing.T) {
	m := Match{
		Host: "example.com",
		Path: "/path",
	}
	upstream := &Upstream{
		Matches: []Match{m},
		Healthy: true,
	}
	assert.True(t, canService("example.com", "/path/resource", upstream))
}

func TestCanServiceRequest_HealthyHostNoPathMatch(t *testing.T) {
	m := Match{
		Host: "example.com",
		Path: "/path",
	}
	upstream := &Upstream{
		Matches: []Match{m},
		Healthy: true,
	}
	assert.False(t, canService("example.com", "/otherpath", upstream))
}

func TestCanServiceRequest_UnhealthyUpstream(t *testing.T) {
	m := Match{
		Host: "example.com",
		Path: "",
	}
	upstream := &Upstream{
		Matches: []Match{m},
		Healthy: false,
	}
	assert.False(t, canService("example.com", "/", upstream))
}

func TestCanServiceRequest_NoHostMatch(t *testing.T) {
	m := Match{
		Host: "example.com",
		Path: "",
	}
	upstream := &Upstream{
		Matches: []Match{m},
		Healthy: true,
	}
	assert.False(t, canService("notexample.com", "/", upstream))
}

func TestCanServiceRequest_EmptyMatches(t *testing.T) {
	upstream := &Upstream{
		Matches: []Match{},
		Healthy: true,
	}
	assert.False(t, canService("example.com", "/", upstream))
}

func TestCanServiceRequest_MultipleMatches(t *testing.T) {
	m1 := Match{
		Host: "example.com",
		Path: "/path1",
	}
	m2 := Match{
		Host: "example.org",
		Path: "/path2",
	}
	upstream := &Upstream{
		Matches: []Match{m1, m2},
		Healthy: true,
	}
	assert.True(t, canService("example.org", "/path2/resource", upstream))
}

func TestPathMatches_EmptyPath(t *testing.T) {
	m := Match{Path: ""}
	upstream := &Upstream{
		Matches: []Match{m},
		Healthy: true,
	}
	assert.True(t, canService("anyhost.com", "/anypath", upstream))
}

func TestPathMatches_NonEmptyPath(t *testing.T) {
	m := Match{Path: "/specific"}
	upstream := &Upstream{
		Matches: []Match{m},
		Healthy: true,
	}
	assert.True(t, canService("anyhost.com", "/specific/resource", upstream))
	assert.False(t, canService("anyhost.com", "/otherpath", upstream))
}

func TestHostMatches_EmptyHost(t *testing.T) {
	m := Match{Host: ""}
	upstream := &Upstream{
		Matches: []Match{m},
		Healthy: true,
	}
	assert.True(t, canService("anyhost.com", "/anypath", upstream))
}

func TestHostMatches_NonEmptyHost(t *testing.T) {
	m := Match{Host: "example.com"}
	upstream := &Upstream{
		Matches: []Match{m},
		Healthy: true,
	}
	assert.True(t, canService("example.com", "/anypath", upstream))
	assert.False(t, canService("notexample.com", "/anypath", upstream))
}
