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

	"p22194.prrrathm.com/mdb"
	mdbrepo "p22194.prrrathm.com/mdb/repository"
	"p22194.prrrathm.com/users/internal/config"
	"p22194.prrrathm.com/users/internal/handler"
	usersrouter "p22194.prrrathm.com/users/internal/router"
	"p22194.prrrathm.com/users/internal/service"
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
	// ── Config ────────────────────────────────────────────────────────────────
	cfgPath := os.Getenv("USERS_CONFIG")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// ── Logger ────────────────────────────────────────────────────────────────
	log := buildLogger(cfg.Log)

	// ── MongoDB ───────────────────────────────────────────────────────────────
	mongoClient, err := mdb.Connect(ctx, cfg.Mongo.URI)
	if err != nil {
		return fmt.Errorf("mongo: %w", err)
	}
	defer mdb.Disconnect(mongoClient)
	log.Info().Str("uri", cfg.Mongo.URI).Msg("connected to MongoDB")

	db := mongoClient.Database(cfg.Mongo.DB)

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := mdbrepo.NewUserRepo(db)
	sessionRepo := mdbrepo.NewSessionRepo(db)

	if err := userRepo.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure indexes: %w", err)
	}

	// ── Service ───────────────────────────────────────────────────────────────
	authSvc := service.New(
		userRepo,
		sessionRepo,
		[]byte(cfg.JWT.Secret),
		cfg.JWT.AccessTTLDuration,
		cfg.JWT.RefreshTTLDuration,
	)

	// ── Handlers & Router ─────────────────────────────────────────────────────
	authHandler := handler.New(authSvc, log)
	httpHandler := usersrouter.New(authHandler, log)

	// ── HTTP Server ───────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	listenErr := make(chan error, 1)
	go func() {
		log.Info().Str("addr", cfg.Addr).Msg("users service starting")
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
	log.Info().Msg("users service stopped")
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
