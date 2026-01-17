package appDynamics

import (
	"context"
	"net/http"

	"github.com/shrihariharanba/go-gateway/internal/telemetry/providers"
)

type AppDynamicsProvider struct {
	cfg providers.Config
}

func New(cfg providers.Config) providers.TelemetryProvider {
	return &AppDynamicsProvider{cfg: cfg}
}

func (a *AppDynamicsProvider) Init(ctx context.Context) error {
	// Typically would initialize the AppDynamics agent here
	// Example:
	// agent.Init(a.cfg.AppDConfig)
	return nil
}

func (a *AppDynamicsProvider) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Example transaction marker:
		// txn := agent.StartTransaction(r.URL.Path)
		// defer txn.End()

		next.ServeHTTP(w, r)
	})
}

func (a *AppDynamicsProvider) Handler() http.Handler {
	// AppDynamics does not expose metrics via HTTP handler
	return nil
}

func (a *AppDynamicsProvider) Name() string { return "appdynamics" }
