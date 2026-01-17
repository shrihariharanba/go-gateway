package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ReverseProxy returns a reverse proxy handler for the given upstream URL
func ReverseProxy(upstream string) http.Handler {
	target, err := url.Parse(upstream)
	if err != nil {
		panic(err)
	}
	return httputil.NewSingleHostReverseProxy(target)
}
