package okta

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/shrihariharanba/go-gateway/internal/sso/providers"
	"golang.org/x/oauth2"
)

type OktaProvider struct {
	cfg        providers.Config
	verifier   *oidc.IDTokenVerifier
	oauth2Conf *oauth2.Config
	issuer     string
}

func NewOktaProvider(cfg providers.Config) (providers.SSOProvider, error) {
	if cfg.ClientID == "" {
		return nil, errors.New("okta sso config missing client_id")
	}
	if cfg.TenantID == "" && cfg.IssuerURL == "" {
		return nil, errors.New("okta sso requires issuer_url or tenant_id")
	}
	if cfg.RedirectURL == "" {
		return nil, errors.New("okta sso config missing redirect_url")
	}

	var issuer string
	if cfg.IssuerURL != "" {
		issuer = cfg.IssuerURL
	} else {
		issuer = fmt.Sprintf("https://%s.okta.com/oauth2/default", cfg.TenantID)
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Okta OIDC provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.ClientID,
	})

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &OktaProvider{
		cfg:        cfg,
		verifier:   verifier,
		oauth2Conf: oauthConfig,
		issuer:     issuer,
	}, nil
}

func (o *OktaProvider) Authenticate(ctx context.Context, token string) (*providers.AuthContext, error) {
	idToken, err := o.verifier.Verify(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to verify Okta ID token: %w", err)
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Subject       string `json:"sub"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse Okta claims: %w", err)
	}

	return &providers.AuthContext{
		UserID:    claims.Subject,
		UserEmail: claims.Email,
		Roles:     []string{"okta-user"},
		Token:     token,
	}, nil
}

func (o *OktaProvider) GetLoginURL() string {
	return o.oauth2Conf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

func (o *OktaProvider) Name() string {
	return "okta"
}
