package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v2"
)

// Config is the top-level kitchen service configuration.
type Config struct {
	Addr     string      `yaml:"addr"`
	GRPCAddr string      `yaml:"grpc_addr"`
	Mongo    MongoConfig `yaml:"mongo"`
	JWT      JWTConfig   `yaml:"jwt"`
	Log      LogConfig   `yaml:"log"`
}

// MongoConfig holds MongoDB connection settings.
type MongoConfig struct {
	URI string `yaml:"uri"`
	DB  string `yaml:"db"`
}

// JWTConfig holds the shared JWT secret used to verify tokens issued by the
// users service. The kitchen service never issues tokens — it only parses them.
type JWTConfig struct {
	Secret string `yaml:"secret"`
}

// LogConfig controls log verbosity and output format.
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func defaults() Config {
	return Config{
		Addr:     ":8083",
		GRPCAddr: ":8084",
		Mongo: MongoConfig{
			URI: "mongodb://localhost:27017",
			DB:  "p22194",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

// Load reads the YAML file at path (if it exists) and overlays KITCHEN_* env vars.
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
		return nil, fmt.Errorf("config: KITCHEN_JWT_SECRET is required")
	}

	return &cfg, nil
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("KITCHEN_ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := os.Getenv("KITCHEN_GRPC_ADDR"); v != "" {
		cfg.GRPCAddr = v
	}
	if v := os.Getenv("KITCHEN_MONGO_URI"); v != "" {
		cfg.Mongo.URI = v
	}
	if v := os.Getenv("KITCHEN_MONGO_DB"); v != "" {
		cfg.Mongo.DB = v
	}
	if v := os.Getenv("KITCHEN_JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("KITCHEN_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
}
