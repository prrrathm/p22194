package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"

	"p22194.prrrathm.com/api-gateway/internal/config"
	internaltls "p22194.prrrathm.com/api-gateway/internal/tls"
)

// Server wraps an http.Server and manages its lifecycle.
type Server struct {
	httpServer *http.Server
	cfg        *config.Config
	log        zerolog.Logger
}

// New constructs a Server from the provided config, handler, and logger.
// TLS is enabled automatically when cfg.TLS.Enabled is true.
func New(cfg *config.Config, handler http.Handler, log zerolog.Logger) (*Server, error) {
	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	if cfg.TLS.Enabled {
		tlsCfg, err := internaltls.New(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("server: tls: %w", err)
		}
		srv.TLSConfig = tlsCfg
	}

	return &Server{httpServer: srv, cfg: cfg, log: log}, nil
}

// Run starts listening for HTTP(S) requests and blocks until ctx is cancelled,
// then performs a graceful shutdown within cfg.Server.ShutdownTimeout.
func (s *Server) Run(ctx context.Context) error {
	listenErr := make(chan error, 1)

	go func() {
		s.log.Info().Str("addr", s.cfg.Server.Addr).Bool("tls", s.cfg.TLS.Enabled).Msg("gateway starting")

		var err error
		if s.cfg.TLS.Enabled {
			// TLSConfig already has the certificate loaded; pass empty strings.
			err = s.httpServer.ListenAndServeTLS("", "")
		} else {
			err = s.httpServer.ListenAndServe()
		}

		if !errors.Is(err, http.ErrServerClosed) {
			listenErr <- err
		}
		close(listenErr)
	}()

	select {
	case err := <-listenErr:
		return fmt.Errorf("server: listen: %w", err)
	case <-ctx.Done():
		s.log.Info().Msg("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server: shutdown: %w", err)
	}

	s.log.Info().Msg("gateway stopped")
	return nil
}
