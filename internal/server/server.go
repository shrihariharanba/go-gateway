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

	"github.com/shrihariharanba/go-gateway/internal/auth"
	"github.com/shrihariharanba/go-gateway/internal/config"
)

type Server struct {
	router      *chi.Mux
	cfg         *config.Config
	oidcHandler *auth.OIDCProvider
	httpServer  *http.Server
}

// NewServer initializes router, OIDC, and routes
func NewServer(cfg *config.Config) *Server {
	r := chi.NewRouter()

	// Health endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	s := &Server{
		router: r,
		cfg:    cfg,
	}

	// Initialize OIDC if enabled
	if cfg.Azure.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		oidcProv, err := auth.NewOIDCProvider(ctx, cfg.Azure.Issuer, cfg.Azure.ClientID)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize OIDC provider")
		}
		s.oidcHandler = oidcProv
		log.Info().Msg("Azure SSO enabled: OIDC provider ready")
	} else {
		log.Warn().Msg("Azure SSO disabled: routes running without auth")
	}

	s.registerRoutes()
	return s
}

// registerRoutes sets up routes with optional OIDC middleware
func (s *Server) registerRoutes() {
	for _, rt := range s.cfg.Routes {
		route := rt

		var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			s.handleReverseProxy(route, w, r)
		}

		// Wrap with OIDC middleware if enabled and route requires auth
		if s.cfg.Azure.Enabled && route.AuthPolicy != "none" {
			authRequired := route.AuthPolicy == "required"
			handler = auth.AuthMiddleware(s.oidcHandler, authRequired)(handler).ServeHTTP
		}

		s.router.HandleFunc(route.Path, handler)
	}
}

// handleReverseProxy forwards request to upstream
func (s *Server) handleReverseProxy(route config.RouteConfig, w http.ResponseWriter, r *http.Request) {
	log.Info().
		Str("method", r.Method).
		Str("path", route.Path).
		Str("upstream", route.Upstream).
		Str("authPolicy", route.AuthPolicy).
		Msg("Proxying request")

	ReverseProxy(route.Upstream).ServeHTTP(w, r)
}

// Start launches the HTTP server with graceful shutdown
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
		Bool("ssoEnabled", s.cfg.Azure.Enabled).
		Msg("Starting gateway server")

	// Channel to listen for termination signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Run server in a goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		if s.cfg.Server.TLSEnabled {
			serverErrChan <- s.httpServer.ListenAndServeTLS("cert.pem", "key.pem")
		} else {
			serverErrChan <- s.httpServer.ListenAndServe()
		}
	}()

	// Wait for signal or server error
	select {
	case sig := <-stopChan:
		log.Info().Str("signal", sig.String()).Msg("Shutdown signal received")
	case err := <-serverErrChan:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Info().Msg("Shutting down server gracefully...")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	log.Info().Msg("Server shutdown completed")
	return nil
}
