package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"

	"p22194.prrrathm.com/api-gateway/internal/cache"
	"p22194.prrrathm.com/api-gateway/internal/config"
	"p22194.prrrathm.com/api-gateway/internal/geoip"
	"p22194.prrrathm.com/api-gateway/internal/health"
	"p22194.prrrathm.com/api-gateway/internal/metrics"
	"p22194.prrrathm.com/api-gateway/internal/proxy"
	"p22194.prrrathm.com/api-gateway/internal/router"
	"p22194.prrrathm.com/api-gateway/server"
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
	// ── Config ──────────────────────────────────────────────────────────────
	cfgPath := os.Getenv("GATEWAY_CONFIG")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// ── Logger ───────────────────────────────────────────────────────────────
	log := buildLogger(cfg.Log)

	// ── Tracer ───────────────────────────────────────────────────────────────
	// No-op tracer provider by default; swap in an OTLP exporter as needed.
	tp := trace.NewTracerProvider()
	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(propagator)
	tracer := tp.Tracer("api-gateway")

	// ── Metrics ──────────────────────────────────────────────────────────────
	metricsRegistry := metrics.New()

	// ── Cache ────────────────────────────────────────────────────────────────
	maxCost := cfg.Cache.MaxCostMB * 1024 * 1024
	responseCache, err := cache.New(maxCost, cfg.Cache.DefaultTTL)
	if err != nil {
		return fmt.Errorf("init cache: %w", err)
	}
	defer responseCache.Close()

	// ── GeoIP (optional) ─────────────────────────────────────────────────────
	var geoReader *geoip.Reader
	if cfg.GeoIP.DBPath != "" {
		geoReader, err = geoip.Open(cfg.GeoIP.DBPath)
		if err != nil {
			log.Warn().Err(err).Msg("geoip database unavailable; geo lookups disabled")
		} else {
			defer geoReader.Close()
		}
	}
	// geoReader may be nil — downstream code must handle that.
	_ = geoReader

	// ── Proxy ────────────────────────────────────────────────────────────────
	proxyHandler, err := proxy.New(cfg.Upstreams, responseCache, metricsRegistry, log)
	if err != nil {
		return fmt.Errorf("init proxy: %w", err)
	}

	// ── Health ───────────────────────────────────────────────────────────────
	healthHandler := health.New(cfg.Upstreams)

	// ── Router ───────────────────────────────────────────────────────────────
	handler, err := router.New(router.Deps{
		Config:     cfg,
		Log:        log,
		Metrics:    metricsRegistry,
		Proxy:      proxyHandler,
		Health:     healthHandler,
		Tracer:     tracer,
		Propagator: propagator,
	})
	if err != nil {
		return fmt.Errorf("init router: %w", err)
	}

	// ── Server ────────────────────────────────────────────────────────────────
	srv, err := server.New(cfg, handler, log)
	if err != nil {
		return fmt.Errorf("init server: %w", err)
	}

	return srv.Run(ctx)
}

// buildLogger creates a zerolog.Logger from the log config.
func buildLogger(cfg config.LogConfig) zerolog.Logger {
	var log zerolog.Logger

	switch cfg.Format {
	case "pretty":
		log = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
		})
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
