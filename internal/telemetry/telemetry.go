package telemetry

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shrihariharanba/go-gateway/internal/telemetry/providers"
	"github.com/shrihariharanba/go-gateway/internal/telemetry/providers/appDynamics"
	"github.com/shrihariharanba/go-gateway/internal/telemetry/providers/newRelic"
	"github.com/shrihariharanba/go-gateway/internal/telemetry/providers/openTelemetry"
	"github.com/shrihariharanba/go-gateway/internal/telemetry/providers/prometheus"
)

type Config struct {
	Providers []providers.Config `yaml:"providers"`
}

type Telemetry struct {
	providers []providers.TelemetryProvider
}

func New(cfg Config) (*Telemetry, error) {
	var list []providers.TelemetryProvider

	for _, pCfg := range cfg.Providers {
		p, err := NewProvider(pCfg)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}

	ctx := context.Background()
	for _, p := range list {
		if err := p.Init(ctx); err != nil {
			return nil, fmt.Errorf("telemetry init failed (%s): %w", p.Name(), err)
		}
	}

	return &Telemetry{providers: list}, nil
}

func (t *Telemetry) Middleware(next http.Handler) http.Handler {
	h := next
	for _, p := range t.providers {
		h = p.Middleware(h)
	}
	return h
}

func (t *Telemetry) RegisterHandlers(r chi.Router) {
	for _, p := range t.providers {
		if h := p.Handler(); h != nil {
			r.Handle("/"+p.Name(), h)
		}
	}
}

func NewProvider(cfg providers.Config) (providers.TelemetryProvider, error) {
	if !cfg.Enabled {
		return providers.NewNoopProvider(), nil
	}

	switch cfg.Type {
	case providers.ProviderPrometheus:
		return prometheus.New(cfg), nil
	case providers.ProviderNewRelic:
		return newRelic.New(cfg), nil
	case providers.ProviderAppD:
		return appDynamics.New(cfg), nil
	case providers.ProviderOTel:
		return openTelemetry.New(cfg), nil
	case providers.ProviderNone:
		return providers.NewNoopProvider(), nil
	default:
		return nil, fmt.Errorf("unknown telemetry provider: %s", cfg.Type)
	}
}
