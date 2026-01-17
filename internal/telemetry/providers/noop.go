package providers

import (
	"context"
	"net/http"
)

type NoopProvider struct{}

func NewNoopProvider() *NoopProvider {
	return &NoopProvider{}
}

func (n *NoopProvider) Init(ctx context.Context) error            { return nil }
func (n *NoopProvider) Middleware(next http.Handler) http.Handler { return next }
func (n *NoopProvider) Handler() http.Handler                     { return nil }
func (n *NoopProvider) Name() string                              { return "noop" }
