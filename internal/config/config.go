package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port       int  `yaml:"port"`
	TLSEnabled bool `yaml:"tlsEnabled"`
}

// AzureConfig holds Azure AD OIDC settings (optional).
type AzureConfig struct {
	Enabled  bool   `yaml:"enabled"`
	TenantID string `yaml:"tenantId"`
	ClientID string `yaml:"clientId"`
	Issuer   string `yaml:"issuer"`
}

// RouteConfig defines a route and upstream target.
type RouteConfig struct {
	Path       string   `yaml:"path"`
	Upstream   string   `yaml:"upstream"`
	Scopes     []string `yaml:"scopes"`
	AuthPolicy string   `yaml:"authPolicy"` // future use: "required" / "optional" / "none"
}

// Config is the root configuration struct.
type Config struct {
	Server ServerConfig  `yaml:"server"`
	Azure  AzureConfig   `yaml:"azure"`
	Routes []RouteConfig `yaml:"routes"`
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

	// Azure validation is conditional based on `Enabled`
	if c.Azure.Enabled {
		if c.Azure.TenantID == "" {
			return errors.New("azure.tenantId is required when azure.enabled=true")
		}
		if c.Azure.ClientID == "" {
			return errors.New("azure.clientId is required when azure.enabled=true")
		}
		if c.Azure.Issuer == "" {
			return errors.New("azure.issuer is required when azure.enabled=true")
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
		p, err := strconv.Atoi(val)
		if err == nil {
			cfg.Server.Port = p
		}
	}
	if val := os.Getenv("GATEWAY_TLS_ENABLED"); val != "" {
		cfg.Server.TLSEnabled = val == "true" || val == "1"
	}

	// Azure overrides
	if val := os.Getenv("AZURE_ENABLED"); val != "" {
		cfg.Azure.Enabled = val == "true" || val == "1"
	}
	if val := os.Getenv("AZURE_TENANT_ID"); val != "" {
		cfg.Azure.TenantID = val
	}
	if val := os.Getenv("AZURE_CLIENT_ID"); val != "" {
		cfg.Azure.ClientID = val
	}
	if val := os.Getenv("AZURE_ISSUER"); val != "" {
		cfg.Azure.Issuer = val
	}
}
