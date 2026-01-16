package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCProvider holds OIDC discovery info and verifier
type OIDCProvider struct {
	Verifier *oidc.IDTokenVerifier
	Config   *oauth2.Config
}

// NewOIDCProvider initializes OIDC provider and verifier
func NewOIDCProvider(ctx context.Context, issuer, clientID string) (*OIDCProvider, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})

	// OAuth2 config placeholder (can be used for code flow / token exchange)
	conf := &oauth2.Config{
		ClientID: clientID,
		Endpoint: provider.Endpoint(),
	}

	return &OIDCProvider{
		Verifier: verifier,
		Config:   conf,
	}, nil
}

// ValidateToken verifies an incoming JWT ID token string
func (p *OIDCProvider) ValidateToken(ctx context.Context, rawToken string) (*oidc.IDToken, error) {
	idToken, err := p.Verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}
	return idToken, nil
}

// ExtractClaims extracts claims into a generic map
func ExtractClaims(idToken *oidc.IDToken) (map[string]interface{}, error) {
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}
	return claims, nil
}

// Simple context key for storing claims
type contextKey string

const ClaimsKey contextKey = "claims"
