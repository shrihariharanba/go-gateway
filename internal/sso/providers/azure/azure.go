package azure

import (
	"context"

	"github.com/shrihariharanba/go-gateway/internal/sso/providers"
)

type AzureProvider struct {
	cfg providers.Config
}

func NewAzureProvider(cfg providers.Config) (providers.SSOProvider, error) {
	return &AzureProvider{cfg: cfg}, nil
}

func (a *AzureProvider) Authenticate(ctx context.Context, token string) (*providers.AuthContext, error) {
	return &providers.AuthContext{
		UserID:    "azure-user",
		UserEmail: "user@azure.com",
		Roles:     []string{"admin"},
		Token:     token,
	}, nil
}

func (a *AzureProvider) GetLoginURL() string {
	return "https://login.microsoftonline.com/"
}

func (a *AzureProvider) Name() string {
	return "azure"
}
