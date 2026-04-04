package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.yaml.in/yaml/v2"
)

// Config is the top-level gateway configuration.
type Config struct {
	Server    ServerConfig              `yaml:"server"`
	Upstreams map[string]UpstreamConfig `yaml:"upstreams"`
	Auth      AuthConfig                `yaml:"auth"`
	RateLimit RateLimitConfig           `yaml:"ratelimit"`
	TLS       TLSConfig                 `yaml:"tls"`
	Log       LogConfig                 `yaml:"log"`
	Metrics   MetricsConfig             `yaml:"metrics"`
	Cache     CacheConfig               `yaml:"cache"`
	GeoIP     GeoIPConfig               `yaml:"geoip"`
	CORS      CORSConfig                `yaml:"cors"`
}

type ServerConfig struct {
	Addr            string        `yaml:"addr"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type UpstreamConfig struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

type AuthConfig struct {
	JWTSecret  string   `yaml:"jwt_secret"`
	SkipRoutes []string `yaml:"skip_routes"`
}

type RateLimitConfig struct {
	RequestsPerSecond float64 `yaml:"requests_per_second"`
	Burst             int     `yaml:"burst"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type LogConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, pretty
}

type MetricsConfig struct {
	Path string `yaml:"path"`
}

type CacheConfig struct {
	MaxCostMB  int64         `yaml:"max_cost_mb"`
	DefaultTTL time.Duration `yaml:"default_ttl"`
}

type GeoIPConfig struct {
	DBPath string `yaml:"db_path"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

// defaults returns a Config pre-populated with sane defaults.
func defaults() Config {
	return Config{
		Server: ServerConfig{
			Addr:            ":8080",
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 15 * time.Second,
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: 100,
			Burst:             200,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		Metrics: MetricsConfig{
			Path: "/metrics",
		},
		Cache: CacheConfig{
			MaxCostMB:  64,
			DefaultTTL: 30 * time.Second,
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
			AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		},
	}
}

// Load reads the YAML file at path (if it exists) and overlays any
// GATEWAY_-prefixed environment variables on top.
//
// Supported env var overrides:
//
//	GATEWAY_SERVER_ADDR             → cfg.Server.Addr
//	GATEWAY_AUTH_JWT_PUBLIC_KEY_PEM → cfg.Auth.JWTPublicKeyPEM
//	GATEWAY_TLS_ENABLED             → cfg.TLS.Enabled
//	GATEWAY_LOG_LEVEL       → cfg.Log.Level
func Load(path string) (*Config, error) {
	cfg := defaults()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("config: read %s: %w", path, err)
		}
		if err == nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("config: parse %s: %w", path, err)
			}
		}
	}

	applyEnv(&cfg)
	return &cfg, nil
}

// applyEnv overlays GATEWAY_* environment variables onto cfg.
func applyEnv(cfg *Config) {
	if v := env("GATEWAY_SERVER_ADDR"); v != "" {
		cfg.Server.Addr = v
	}
	if v := env("GATEWAY_LOG_LEVEL"); v != "" {
		cfg.Log.Level = strings.ToLower(v)
	}
	if v := env("GATEWAY_LOG_FORMAT"); v != "" {
		cfg.Log.Format = strings.ToLower(v)
	}
	if v := env("GATEWAY_AUTH_JWT_SECRET"); v != "" {
		cfg.Auth.JWTSecret = v
	}
	if v := env("GATEWAY_METRICS_PATH"); v != "" {
		cfg.Metrics.Path = v
	}
	if v := env("GATEWAY_GEOIP_DB_PATH"); v != "" {
		cfg.GeoIP.DBPath = v
	}
	if v := env("GATEWAY_TLS_CERT_FILE"); v != "" {
		cfg.TLS.CertFile = v
	}
	if v := env("GATEWAY_TLS_KEY_FILE"); v != "" {
		cfg.TLS.KeyFile = v
	}
}

func env(key string) string { return os.Getenv(key) }
