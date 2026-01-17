package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	ssoProviders "github.com/shrihariharanba/go-gateway/internal/sso/providers"
	telemetryProviders "github.com/shrihariharanba/go-gateway/internal/telemetry/providers"
	"gopkg.in/yaml.v3"
)

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port       int  `yaml:"port"`
	TLSEnabled bool `yaml:"tlsEnabled"`
}

// SSOConfig holds generic SSO settings for all providers.
type SSOConfig struct {
	Enabled      bool                      `yaml:"enabled"`
	Type         ssoProviders.ProviderType `yaml:"type"` // none, azure, google, okta
	ClientID     string                    `yaml:"clientId"`
	ClientSecret string                    `yaml:"clientSecret"`
	TenantID     string                    `yaml:"tenantId"`  // Azure or Okta
	IssuerURL    string                    `yaml:"issuerUrl"` // Okta preferred or Azure optional
	RedirectURL  string                    `yaml:"redirectUrl"`
}

// TelemetryConfig holds generic telemetry settings.
type TelemetryConfig struct {
	Enabled  bool                            `yaml:"enabled"`
	Type     telemetryProviders.ProviderType `yaml:"type"`     // prometheus, otel, newrelic, appdynamics
	Endpoint string                          `yaml:"endpoint"` // OTEL / AppDynamics / NewRelic endpoint
	APIKey   string                          `yaml:"apiKey"`   // NR / AppDynamics key
	PromPath string                          `yaml:"promPath"` // Prometheus metrics path
	Service  string                          `yaml:"service"`  // optional service name
}

// RouteConfig defines a route and upstream target.
type RouteConfig struct {
	Path       string   `yaml:"path"`
	Upstream   string   `yaml:"upstream"`
	Scopes     []string `yaml:"scopes"`
	AuthPolicy string   `yaml:"authPolicy"` // "required" / "optional" / "none"
}

// Config is the root configuration struct.
type Config struct {
	Server    ServerConfig      `yaml:"server"`
	SSO       SSOConfig         `yaml:"sso"`
	Telemetry []TelemetryConfig `yaml:"telemetry"`
	Routes    []RouteConfig     `yaml:"routes"`
}

// Load reads YAML config from a file path and applies env overrides.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	applyEnvOverrides(&cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate ensures required fields are set consistently.
func (c *Config) Validate() error {
	if c.Server.Port == 0 {
		return errors.New("server.port must be set")
	}

	// SSO validation
	if c.SSO.Enabled {
		if c.SSO.Type == ssoProviders.ProviderNone {
			return errors.New("sso.type cannot be 'none' if sso.enabled=true")
		}
		if c.SSO.ClientID == "" {
			return errors.New("sso.clientId is required when sso.enabled=true")
		}
		if c.SSO.ClientSecret == "" {
			return errors.New("sso.clientSecret is required when sso.enabled=true")
		}
		if c.SSO.RedirectURL == "" {
			return errors.New("sso.redirectUrl is required when sso.enabled=true")
		}
		// Provider-specific checks
		switch c.SSO.Type {
		case ssoProviders.ProviderAzure:
			if c.SSO.TenantID == "" {
				return errors.New("sso.tenantId is required for Azure")
			}
		case ssoProviders.ProviderOkta:
			if c.SSO.IssuerURL == "" && c.SSO.TenantID == "" {
				return errors.New("sso.issuerUrl or sso.tenantId is required for Okta")
			}
		}
	}

	// Telemetry validation
	for _, t := range c.Telemetry {
		if !t.Enabled {
			continue
		}
		if t.Type == telemetryProviders.ProviderNone {
			return errors.New("telemetry.type cannot be 'none' if telemetry.enabled=true")
		}
		switch t.Type {
		case telemetryProviders.ProviderPrometheus:
			if t.PromPath == "" {
				return errors.New("telemetry.promPath is required for Prometheus")
			}
		case telemetryProviders.ProviderOTel, telemetryProviders.ProviderNewRelic, telemetryProviders.ProviderAppD:
			if t.Endpoint == "" {
				return fmt.Errorf("telemetry.endpoint is required for %s", t.Type)
			}
		}
	}

	// Route validation
	for _, r := range c.Routes {
		if r.Path == "" {
			return errors.New("each route must have a path")
		}
		if r.Upstream == "" {
			return fmt.Errorf("route '%s' must have an upstream", r.Path)
		}
	}

	return nil
}

// applyEnvOverrides allows ENV vars to override config fields.
func applyEnvOverrides(cfg *Config) {
	// Server overrides
	if val := os.Getenv("GATEWAY_PORT"); val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			cfg.Server.Port = p
		}
	}
	if val := os.Getenv("GATEWAY_TLS_ENABLED"); val != "" {
		cfg.Server.TLSEnabled = val == "true" || val == "1"
	}

	// SSO overrides
	if val := os.Getenv("SSO_ENABLED"); val != "" {
		cfg.SSO.Enabled = val == "true" || val == "1"
	}
	if val := os.Getenv("SSO_TYPE"); val != "" {
		cfg.SSO.Type = ssoProviders.ProviderType(val)
	}
	if val := os.Getenv("SSO_CLIENT_ID"); val != "" {
		cfg.SSO.ClientID = val
	}
	if val := os.Getenv("SSO_CLIENT_SECRET"); val != "" {
		cfg.SSO.ClientSecret = val
	}
	if val := os.Getenv("SSO_REDIRECT_URL"); val != "" {
		cfg.SSO.RedirectURL = val
	}
	if val := os.Getenv("SSO_TENANT_ID"); val != "" {
		cfg.SSO.TenantID = val
	}
	if val := os.Getenv("SSO_ISSUER_URL"); val != "" {
		cfg.SSO.IssuerURL = val
	}

	// Telemetry overrides
	for i := range cfg.Telemetry {
		prefix := fmt.Sprintf("TELEMETRY_%d_", i)
		t := &cfg.Telemetry[i]

		if val := os.Getenv(prefix + "ENABLED"); val != "" {
			t.Enabled = val == "true" || val == "1"
		}
		if val := os.Getenv(prefix + "TYPE"); val != "" {
			t.Type = telemetryProviders.ProviderType(val)
		}
		if val := os.Getenv(prefix + "ENDPOINT"); val != "" {
			t.Endpoint = val
		}
		if val := os.Getenv(prefix + "APIKEY"); val != "" {
			t.APIKey = val
		}
		if val := os.Getenv(prefix + "PROMPATH"); val != "" {
			t.PromPath = val
		}
		if val := os.Getenv(prefix + "SERVICE"); val != "" {
			t.Service = val
		}
	}
}
