package main

import (
	"fmt"
	"github.com/maddalax/multiproxy"
	"net/http"
	"net/url"
	"time"
)

type UpstreamMeta struct {
	SomeField any
}

func main() {
	lb := multiproxy.CreateLoadBalancer[UpstreamMeta]()

	go func() {
		for {
			lb.PrintMetrics()
			time.Sleep(3 * time.Second)
		}
	}()

	lb.BeforeRequest = func(up *multiproxy.Upstream[UpstreamMeta], req *http.Request) {
		req.Header.Add("X-Forwarded-For", "")
	}

	lb.AfterRequest = func(up *multiproxy.Upstream[UpstreamMeta], req *http.Request, res *http.Response) {
	}

	lb.OnError = func(up *multiproxy.Upstream[UpstreamMeta], req *http.Request, err error) {
	}

	lb.OnMarkUnhealthy = func(up *multiproxy.Upstream[UpstreamMeta]) {
		//slog.Info("Upstream marked as unhealthy", slog.String("host", up.Url.Host))
	}

	lb.OnMarkHealthy = func(up *multiproxy.Upstream[UpstreamMeta]) {
		//slog.Info("Upstream marked as healthy", slog.String("host", up.Url.Host))
	}

	upstreamUrl, _ := url.Parse("http://localhost:4001")
	upstreamUrl2, _ := url.Parse("http://localhost:4000")
	upstreamUrl3, _ := url.Parse("http://localhost:4003")

	lb.Add(&multiproxy.Upstream[UpstreamMeta]{
		Url:     upstreamUrl,
		Healthy: false,
		Matches: []multiproxy.Match{
			{
				Host: "paas.htmgo.dev",
				Path: "",
			},
		},
		Metadata: UpstreamMeta{
			SomeField: "some value",
		},
	})

	lb.Add(&multiproxy.Upstream[UpstreamMeta]{
		Url:     upstreamUrl2,
		Healthy: false,
		Matches: []multiproxy.Match{
			{
				Host: "paas.htmgo.dev",
				Path: "",
			},
		},
		Metadata: UpstreamMeta{
			SomeField: "some value",
		},
	})

	lb.Add(&multiproxy.Upstream[UpstreamMeta]{
		Url:     upstreamUrl3,
		Healthy: false,
		Matches: []multiproxy.Match{
			{
				Host: "paas.htmgo.dev",
				Path: "",
			},
		},
		Metadata: UpstreamMeta{
			SomeField: "some value",
		},
	})

	for i := 0; i < 20; i++ {
		u, _ := url.Parse(fmt.Sprintf("http://localhost:400%d", i))
		lb.Add(&multiproxy.Upstream[UpstreamMeta]{
			Url:     u,
			Healthy: false,
			Matches: []multiproxy.Match{
				{
					Host: "paas.htmgo.dev",
					Path: "",
				},
			},
		})
	}

	handler := multiproxy.NewReverseProxyHandler(lb)

	http.HandleFunc("/", handler)
	http.ListenAndServe(":80", nil)
}
