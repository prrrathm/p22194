package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"p22194.prrrathm.com/mdb"
	"p22194.prrrathm.com/kitchen/internal/collab"
	"p22194.prrrathm.com/kitchen/internal/config"
	"p22194.prrrathm.com/kitchen/internal/handler"
	"p22194.prrrathm.com/kitchen/internal/repository"
	kitchenrouter "p22194.prrrathm.com/kitchen/internal/router"
	"p22194.prrrathm.com/kitchen/internal/service"
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
	cfgPath := os.Getenv("KITCHEN_CONFIG")
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
	docRepo := repository.NewDocumentRepo(db)
	blockRepo := repository.NewBlockRepo(db)
	sharingRepo := repository.NewSharingRepo(db)

	for _, idx := range []interface{ EnsureIndexes(context.Context) error }{
		docRepo, blockRepo, sharingRepo,
	} {
		if err := idx.EnsureIndexes(ctx); err != nil {
			return fmt.Errorf("ensure indexes: %w", err)
		}
	}

	// ── Services ──────────────────────────────────────────────────────────────
	docSvc := service.NewDocumentService(docRepo, blockRepo)
	blockSvc := service.NewBlockService(blockRepo, docSvc)
	sharingSvc := service.NewSharingService(sharingRepo)

	// ── HTTP handlers & router ────────────────────────────────────────────────
	docHandler := handler.NewDocumentHandler(docSvc, log)
	blockHandler := handler.NewBlockHandler(blockSvc, log)
	sharingHandler := handler.NewSharingHandler(sharingSvc, log)

	httpHandler := kitchenrouter.New(
		docHandler,
		blockHandler,
		sharingHandler,
		[]byte(cfg.JWT.Secret),
		log,
	)

	// ── gRPC collab server ────────────────────────────────────────────────────
	hub := collab.NewHub()
	collabSrv := collab.NewServer(hub, blockRepo, log)

	grpcSrv := grpc.NewServer()
	collab.RegisterCollabServiceServer(grpcSrv, collabSrv)

	grpcLis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}

	// ── HTTP server ───────────────────────────────────────────────────────────
	httpSrv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── Start servers ─────────────────────────────────────────────────────────
	errs := make(chan error, 2)

	go func() {
		log.Info().Str("addr", cfg.Addr).Msg("kitchen HTTP service starting")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errs <- fmt.Errorf("http: %w", err)
		}
	}()

	go func() {
		log.Info().Str("addr", cfg.GRPCAddr).Msg("kitchen gRPC collab service starting")
		if err := grpcSrv.Serve(grpcLis); err != nil {
			errs <- fmt.Errorf("grpc: %w", err)
		}
	}()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
		log.Info().Msg("shutdown signal received")
	}

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	grpcSrv.GracefulStop()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}

	log.Info().Msg("kitchen service stopped")
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
