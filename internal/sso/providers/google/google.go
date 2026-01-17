package google

import (
	"context"

	"github.com/shrihariharanba/go-gateway/internal/sso/providers"
)

type GoogleProvider struct {
	cfg providers.Config
}

func NewGoogleProvider(cfg providers.Config) (providers.SSOProvider, error) {
	return &GoogleProvider{cfg: cfg}, nil
}

func (g *GoogleProvider) Authenticate(ctx context.Context, token string) (*providers.AuthContext, error) {
	return &providers.AuthContext{
		UserID:    "google-user",
		UserEmail: "user@gmail.com",
		Roles:     []string{"user"},
		Token:     token,
	}, nil
}

func (g *GoogleProvider) GetLoginURL() string {
	return "https://accounts.google.com/o/oauth2/v2/auth"
}

func (g *GoogleProvider) Name() string {
	return "google"
}
