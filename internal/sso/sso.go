package sso

import (
	"github.com/shrihariharanba/go-gateway/internal/sso/providers"
	"github.com/shrihariharanba/go-gateway/internal/sso/providers/azure"
	"github.com/shrihariharanba/go-gateway/internal/sso/providers/google"
	"github.com/shrihariharanba/go-gateway/internal/sso/providers/okta"
)

func NewProvider(cfg providers.Config) (providers.SSOProvider, error) {
	if !cfg.Enabled {
		return &NoAuthProvider{}, nil
	}

	switch cfg.Type {
	case providers.ProviderAzure:
		return azure.NewAzureProvider(cfg)
	case providers.ProviderGoogle:
		return google.NewGoogleProvider(cfg)
	case providers.ProviderOkta:
		return okta.NewOktaProvider(cfg)
	default:
		return &NoAuthProvider{}, nil
	}
}
