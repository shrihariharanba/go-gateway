package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/shrihariharanba/go-gateway/internal/server/proxy"

	"github.com/shrihariharanba/go-gateway/internal/config"
	"github.com/shrihariharanba/go-gateway/internal/sso"
	"github.com/shrihariharanba/go-gateway/internal/sso/providers"
	"github.com/shrihariharanba/go-gateway/internal/telemetry"
	teleprovider "github.com/shrihariharanba/go-gateway/internal/telemetry/providers"
)

type Server struct {
	router      *chi.Mux
	cfg         *config.Config
	httpServer  *http.Server
	ssoProvider providers.SSOProvider
	telemetry   *telemetry.Telemetry
}

func NewServer(cfg *config.Config) *Server {
	r := chi.NewRouter()

	// Create server
	s := &Server{
		router: r,
		cfg:    cfg,
	}

	// ---------------------------
	// SSO Provider Setup
	// ---------------------------
	if cfg.SSO.Enabled {
		pCfg := providers.Config{
			Enabled:      cfg.SSO.Enabled,
			Type:         cfg.SSO.Type,
			ClientID:     cfg.SSO.ClientID,
			ClientSecret: cfg.SSO.ClientSecret,
			TenantID:     cfg.SSO.TenantID,
			IssuerURL:    cfg.SSO.IssuerURL,
			RedirectURL:  cfg.SSO.RedirectURL,
		}

		provider, err := sso.NewProvider(pCfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize SSO provider")
		}
		s.ssoProvider = provider
		log.Info().Str("provider", provider.Name()).Msg("SSO enabled")
	} else {
		s.ssoProvider = &sso.NoAuthProvider{}
		log.Warn().Msg("SSO disabled: running without authentication")
	}

	// ---------------------------
	// Telemetry Setup
	// ---------------------------
	if len(cfg.Telemetry) > 0 {
		var provCfgs []teleprovider.Config
		for _, tcfg := range cfg.Telemetry {
			provCfgs = append(provCfgs, teleprovider.Config{
				Enabled:     tcfg.Enabled,
				Type:        tcfg.Type,
				Endpoint:    tcfg.Endpoint,
				APIKey:      tcfg.APIKey,
				PromPath:    tcfg.PromPath,
				ServiceName: tcfg.Service,
			})
		}

		tel, err := telemetry.New(telemetry.Config{Providers: provCfgs})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize telemetry")
		}

		s.telemetry = tel

		// ------ IMPORTANT ------
		// Must apply middleware BEFORE routes
		// -----------------------
		if s.telemetry != nil {
			r.Use(s.telemetry.Middleware)
			s.telemetry.RegisterHandlers(r)
		}
	}

	// ---------------------------
	// Health endpoint (no auth)
	// ---------------------------
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// ---------------------------
	// Register application routes
	// ---------------------------
	s.registerRoutes()

	return s
}

// ----------------------------------------------
// ROUTES
// ----------------------------------------------
func (s *Server) registerRoutes() {
	for _, rt := range s.cfg.Routes {
		route := rt

		var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.handleReverseProxy(route, w, r)
		})

		// SSO per-route policy
		if s.ssoProvider != nil && route.AuthPolicy != "none" {
			authRequired := route.AuthPolicy == "required"
			handler = sso.AuthMiddleware(s.ssoProvider, authRequired)(handler)
		}

		s.router.Handle(route.Path, handler)
	}
}

// ----------------------------------------------
// PROXY
// ----------------------------------------------
func (s *Server) handleReverseProxy(route config.RouteConfig, w http.ResponseWriter, r *http.Request) {
	log.Info().
		Str("method", r.Method).
		Str("path", route.Path).
		Str("upstream", route.Upstream).
		Str("authPolicy", route.AuthPolicy).
		Msg("Proxying request")

	proxy.ReverseProxy(route.Upstream).ServeHTTP(w, r)
}

// ----------------------------------------------
// SERVER START / SHUTDOWN
// ----------------------------------------------
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Server.Port)
	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Info().
		Str("addr", addr).
		Bool("tls", s.cfg.Server.TLSEnabled).
		Bool("sso", s.cfg.SSO.Enabled).
		Msg("Starting gateway server")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	serverErrChan := make(chan error)

	go func() {
		if s.cfg.Server.TLSEnabled {
			serverErrChan <- s.httpServer.ListenAndServeTLS("cert.pem", "key.pem")
		} else {
			serverErrChan <- s.httpServer.ListenAndServe()
		}
	}()

	select {
	case <-stopChan:
		log.Warn().Msg("Received shutdown signal")
	case err := <-serverErrChan:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Info().Msg("Graceful shutdown...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown failed: %w", err)
	}

	log.Info().Msg("Server stopped")
	return nil
}
