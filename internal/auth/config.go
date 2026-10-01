package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds application configuration for auth and mail services
type Config struct {
	SMTPEmail          string `json:"smtp_email"`
	SMTPAppPassword    string `json:"smtp_app_password"`
	SMTPHost           string `json:"smtp_host"`
	SMTPPort           string `json:"smtp_port"`
	JWTSecret          string `json:"jwt_secret"`
	JWTExpirationHours int    `json:"jwt_expiration_hours"`
	DataDir            string `json:"data_dir"`
}

// DefaultConfig returns safe production-ready defaults
func DefaultConfig() *Config {
	return &Config{
		SMTPEmail:          "",
		SMTPAppPassword:    "",
		SMTPHost:           "smtp.gmail.com",
		SMTPPort:           "587",
		JWTSecret:          "rutils-utility-jwt-secret-change-in-production-2026",
		JWTExpirationHours: 24,
		DataDir:            "data",
	}
}

// LoadConfig reads configuration from path or falls back to defaults
func LoadConfig(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	if configPath == "" {
		configPath = "config.json"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Apply sensible fallbacks if empty
	if cfg.SMTPHost == "" {
		cfg.SMTPHost = "smtp.gmail.com"
	}
	if cfg.SMTPPort == "" {
		cfg.SMTPPort = "587"
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "rutils-utility-jwt-secret-change-in-production-2026"
	}
	if cfg.JWTExpirationHours <= 0 {
		cfg.JWTExpirationHours = 24
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "data"
	}

	return cfg, nil
}

// GetUsersFilePath returns the resolved path to data/users.json
func (c *Config) GetUsersFilePath() string {
	return filepath.Join(c.DataDir, "users.json")
}
