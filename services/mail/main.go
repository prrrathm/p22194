package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"p22194.prrrathm.com/mail/internal/config"
	"p22194.prrrathm.com/mail/internal/handler"
	"p22194.prrrathm.com/mail/internal/router"
	"p22194.prrrathm.com/mail/internal/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfgPath := os.Getenv("MAIL_CONFIG")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := buildLogger(cfg.Log)

	mailer := service.NewMailer(
		cfg.Mail.ResendAPIBase,
		cfg.Mail.ResendAPIKey,
		cfg.Mail.FromEmail,
		cfg.Mail.FromName,
		&http.Client{Timeout: cfg.Mail.RequestTimeoutDuration},
	)

	mailHandler := handler.New(mailer, log)
	httpHandler := router.New(mailHandler, log)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	listenErr := make(chan error, 1)
	go func() {
		log.Info().Str("addr", cfg.Addr).Msg("mail service starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			listenErr <- err
		}
		close(listenErr)
	}()

	select {
	case err := <-listenErr:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
		log.Info().Msg("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	log.Info().Msg("mail service stopped")
	return nil
}

func buildLogger(cfg config.LogConfig) zerolog.Logger {
	var log zerolog.Logger
	switch cfg.Format {
	case "pretty":
		log = zerolog.New(zerolog.NewConsoleWriter())
	default:
		log = zerolog.New(os.Stdout)
	}
	log = log.With().Timestamp().Logger()
	switch cfg.Level {
	case "debug":
		log = log.Level(zerolog.DebugLevel)
	case "warn":
		log = log.Level(zerolog.WarnLevel)
	case "error":
		log = log.Level(zerolog.ErrorLevel)
	default:
		log = log.Level(zerolog.InfoLevel)
	}
	return log
}
