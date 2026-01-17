package providers

import (
	"context"
	"net/http"
)

type ProviderType string

const (
	ProviderNone       ProviderType = "none"
	ProviderPrometheus ProviderType = "prometheus"
	ProviderNewRelic   ProviderType = "newrelic"
	ProviderAppD       ProviderType = "appdynamics"
	ProviderOTel       ProviderType = "opentelemetry"
)

type TelemetryProvider interface {
	Init(ctx context.Context) error
	Middleware(next http.Handler) http.Handler
	Handler() http.Handler
	Name() string
}

type Config struct {
	Enabled     bool
	Type        ProviderType
	Endpoint    string
	APIKey      string
	PromPath    string
	ServiceName string
}
