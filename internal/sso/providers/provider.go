package providers

import "context"

type AuthContext struct {
	UserID    string
	UserEmail string
	Roles     []string
	Token     string
}

type SSOProvider interface {
	Authenticate(ctx context.Context, token string) (*AuthContext, error)
	GetLoginURL() string
	Name() string
}

type ProviderType string

const (
	ProviderNone   ProviderType = "none"
	ProviderAzure  ProviderType = "azure"
	ProviderGoogle ProviderType = "google"
	ProviderOkta   ProviderType = "okta"
)

type Config struct {
	Enabled      bool
	Type         ProviderType
	ClientID     string
	ClientSecret string
	TenantID     string
	IssuerURL    string
	RedirectURL  string
}
