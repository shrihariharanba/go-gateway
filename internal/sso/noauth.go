package sso

import (
	"context"

	"github.com/shrihariharanba/go-gateway/internal/sso/providers"
)

type NoAuthProvider struct{}

// Authenticate returns an anonymous user
func (n *NoAuthProvider) Authenticate(ctx context.Context, token string) (*providers.AuthContext, error) {
	return &providers.AuthContext{
		UserID:    "anonymous",
		UserEmail: "",
		Roles:     []string{"public"},
		Token:     token,
	}, nil
}

// GetLoginURL returns a generic login path
func (n *NoAuthProvider) GetLoginURL() string {
	return "/login"
}

// Name returns provider type "none"
func (n *NoAuthProvider) Name() string {
	return "none"
}
