package prometheus

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"context"

	"github.com/shrihariharanba/go-gateway/internal/telemetry/providers"
)

type PromProvider struct {
	cfg      providers.Config
	registry *prometheus.Registry
}

func New(cfg providers.Config) providers.TelemetryProvider {
	return &PromProvider{
		cfg:      cfg,
		registry: prometheus.NewRegistry(),
	}
}

func (p *PromProvider) Init(ctx context.Context) error {
	p.registry.MustRegister(prometheus.NewGoCollector())
	p.registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	return nil
}

func (p *PromProvider) Middleware(next http.Handler) http.Handler {
	reqs := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total requests",
		},
		[]string{"path", "method", "status"},
	)
	p.registry.MustRegister(reqs)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := &responseRecorder{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(rr, r)
		reqs.WithLabelValues(r.URL.Path, r.Method, fmt.Sprintf("%d", rr.statusCode)).Inc()
	})
}

func (p *PromProvider) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}

func (p *PromProvider) Name() string { return "prometheus" }

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}
