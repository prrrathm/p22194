package config

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v2"
)

type Config struct {
	Addr string     `yaml:"addr"`
	Mail MailConfig `yaml:"mail"`
	Log  LogConfig  `yaml:"log"`
}

type MailConfig struct {
	FromEmail      string `yaml:"from_email"`
	FromName       string `yaml:"from_name"`
	ResendAPIKey   string `yaml:"-"`
	ResendAPIBase  string `yaml:"resend_api_base"`
	RequestTimeout string `yaml:"request_timeout"`

	RequestTimeoutDuration time.Duration `yaml:"-"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func defaults() Config {
	return Config{
		Addr: ":8085",
		Mail: MailConfig{
			FromEmail:      "noreply@example.com",
			FromName:       "P22194",
			ResendAPIBase:  "https://api.resend.com",
			RequestTimeout: "10s",
		},
		Log: LogConfig{Level: "info", Format: "pretty"},
	}
}

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

	if cfg.Mail.ResendAPIKey == "" {
		return nil, fmt.Errorf("config: RESEND_API_KEY is required")
	}

	var err error
	cfg.Mail.RequestTimeoutDuration, err = time.ParseDuration(cfg.Mail.RequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("config: invalid mail.request_timeout %q: %w", cfg.Mail.RequestTimeout, err)
	}

	return &cfg, nil
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("MAIL_ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := os.Getenv("MAIL_FROM_EMAIL"); v != "" {
		cfg.Mail.FromEmail = v
	}
	if v := os.Getenv("MAIL_FROM_NAME"); v != "" {
		cfg.Mail.FromName = v
	}
	if v := os.Getenv("MAIL_RESEND_API_BASE"); v != "" {
		cfg.Mail.ResendAPIBase = v
	}
	if v := os.Getenv("MAIL_REQUEST_TIMEOUT"); v != "" {
		cfg.Mail.RequestTimeout = v
	}
	if v := os.Getenv("RESEND_API_KEY"); v != "" {
		cfg.Mail.ResendAPIKey = v
	}
	if v := os.Getenv("MAIL_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv("MAIL_LOG_FORMAT"); v != "" {
		cfg.Log.Format = v
	}
}
