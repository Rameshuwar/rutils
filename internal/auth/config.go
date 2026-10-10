package auth

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Config holds application configuration for auth and mail services
type Config struct {
	SMTPEmail          string `json:"smtp_email"`
	SMTPAppPassword    string `json:"smtp_app_password"`
	SMTPHost           string `json:"smtp_host"`
	SMTPPort           string `json:"smtp_port"`
	SMTPSenderName     string `json:"smtp_sender_name"`
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
		SMTPSenderName:     "Rutils Team",
		JWTSecret:          "rutils-utility-jwt-secret-change-in-production-2026",
		JWTExpirationHours: 24,
		DataDir:            "data",
	}
}

// IsSMTPConfigured returns true if SMTP credentials are fully provided
func (c *Config) IsSMTPConfigured() bool {
	return c != nil &&
		strings.TrimSpace(c.SMTPEmail) != "" &&
		strings.TrimSpace(c.SMTPAppPassword) != ""
}

// LoadConfig reads configuration from path or falls back to defaults
func LoadConfig(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	var searchPaths []string
	if configPath != "" {
		searchPaths = append(searchPaths, configPath)
	}
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		searchPaths = append(searchPaths, envPath)
	}
	searchPaths = append(searchPaths,
		"config.json",
		filepath.Join("rutils", "config.json"),
		filepath.Join("..", "config.json"),
		filepath.Join("..", "..", "config.json"),
	)

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		searchPaths = append(searchPaths, filepath.Join(exeDir, "config.json"), filepath.Join(exeDir, "..", "config.json"))
	}

	var foundPath string
	var data []byte
	for _, p := range searchPaths {
		if content, err := os.ReadFile(p); err == nil {
			foundPath = p
			data = content
			break
		}
	}

	if data == nil {
		log.Printf("[AUTH-WARN] No config.json found in search paths %v. Running in default dev fallback mode (SMTP disabled).", searchPaths)
		return cfg, nil
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
	if cfg.SMTPSenderName == "" {
		cfg.SMTPSenderName = "Rutils Team"
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

	if cfg.IsSMTPConfigured() {
		log.Printf("[AUTH] Loaded configuration from %q: SMTP active for sender <%s> (host: %s:%s)", foundPath, cfg.SMTPEmail, cfg.SMTPHost, cfg.SMTPPort)
	} else {
		log.Printf("[AUTH-WARN] Loaded configuration from %q: SMTP not configured (smtp_email or smtp_app_password missing).", foundPath)
	}

	return cfg, nil
}

// GetDataDir returns resolved data directory path
func (c *Config) GetDataDir() string {
	dir := c.DataDir
	if dir == "" {
		dir = "data"
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		alt := filepath.Join("rutils", dir)
		if fi, err := os.Stat(alt); err == nil && fi.IsDir() {
			return alt
		}
	}
	return dir
}

// GetUsersFilePath returns the resolved path to data/users.json
func (c *Config) GetUsersFilePath() string {
	return filepath.Join(c.GetDataDir(), "users.json")
}
