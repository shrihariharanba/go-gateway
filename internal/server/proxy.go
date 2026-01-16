package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// ReverseProxy returns a configured reverse proxy handler for a given upstream.
func ReverseProxy(upstream string) http.Handler {
	target, err := url.Parse(upstream)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "invalid upstream URL", http.StatusBadGateway)
		})
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Optional: modify request before forwarding
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Ensure host headers are set correctly
		req.Host = target.Host

		// Optional: add forwarded headers
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
	}

	// Timeout & transport settings
	proxy.Transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DisableKeepAlives:     false,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Error handling
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "upstream service unavailable", http.StatusBadGateway)
	}

	return proxy
}
