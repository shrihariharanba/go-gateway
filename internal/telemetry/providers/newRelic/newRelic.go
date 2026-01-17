package newRelic

import (
	"context"
	"net/http"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/shrihariharanba/go-gateway/internal/telemetry/providers"
)

type NewRelicProvider struct {
	cfg providers.Config
	app *newrelic.Application
}

func New(cfg providers.Config) providers.TelemetryProvider {
	return &NewRelicProvider{cfg: cfg}
}

func (n *NewRelicProvider) Init(ctx context.Context) error {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(n.cfg.ServiceName),
		newrelic.ConfigLicense(n.cfg.APIKey),
		newrelic.ConfigEnabled(true),
	)
	if err != nil {
		return err
	}
	n.app = app
	return nil
}

func (n *NewRelicProvider) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		txn := n.app.StartTransaction(r.URL.Path)
		defer txn.End()

		w = txn.SetWebResponse(w)
		txn.SetWebRequestHTTP(r)

		next.ServeHTTP(w, r)
	})
}

func (n *NewRelicProvider) Handler() http.Handler {
	// No direct HTTP endpoint needed for NR
	return nil
}

func (n *NewRelicProvider) Name() string { return "newrelic" }
