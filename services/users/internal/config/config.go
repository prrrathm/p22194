package config

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v2"
	
)

type Config struct {
	Addr  string      `yaml:"addr"`
	Mongo MongoConfig `yaml:"mongo"`
	JWT   JWTConfig   `yaml:"jwt"`
	Log   LogConfig   `yaml:"log"`
}

type MongoConfig struct {
	URI string `yaml:"uri"`
	DB  string `yaml:"db"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`      // no YAML default; env-only
	AccessTTL  string `yaml:"access_ttl"`  // parsed via time.ParseDuration
	RefreshTTL string `yaml:"refresh_ttl"` // parsed via time.ParseDuration

	// Parsed durations — populated by Load after YAML unmarshal.
	AccessTTLDuration  time.Duration `yaml:"-"`
	RefreshTTLDuration time.Duration `yaml:"-"`
}

type LogConfig struct {
Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func defaults() Config {
	return Config{
		Addr: ":8082",
		Mongo: MongoConfig{
			URI: "mongodb://localhost:27017",
			DB:  "p22194",
		},
		JWT: JWTConfig{
			AccessTTL:  "15m",
			RefreshTTL: "168h",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

// Load reads the YAML file at path (if it exists) and overlays USERS_* env vars.
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

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("config: USERS_JWT_SECRET is required")
	}

	var err error
	cfg.JWT.AccessTTLDuration, err = time.ParseDuration(cfg.JWT.AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("config: invalid jwt.access_ttl %q: %w", cfg.JWT.AccessTTL, err)
	}
	cfg.JWT.RefreshTTLDuration, err = time.ParseDuration(cfg.JWT.RefreshTTL)
	if err != nil {
		return nil, fmt.Errorf("config: invalid jwt.refresh_ttl %q: %w", cfg.JWT.RefreshTTL, err)
	}

	return &cfg, nil
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("USERS_ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := os.Getenv("USERS_MONGO_URI"); v != "" {
		cfg.Mongo.URI = v
	}
	if v := os.Getenv("USERS_MONGO_DB"); v != "" {
		cfg.Mongo.DB = v
	}
	if v := os.Getenv("USERS_JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("USERS_JWT_ACCESS_TTL"); v != "" {
		cfg.JWT.AccessTTL = v
	}
	if v := os.Getenv("USERS_JWT_REFRESH_TTL"); v != "" {
		cfg.JWT.RefreshTTL = v
	}
	if v := os.Getenv("USERS_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
}
