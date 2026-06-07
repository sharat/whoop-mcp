package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

const RedirectURI = "http://127.0.0.1:8181/callback"
const AuthURL = "https://api.prod.whoop.com/oauth/oauth2/auth"
const TokenURL = "https://api.prod.whoop.com/oauth/oauth2/token"
const BaseURL = "https://api.prod.whoop.com/developer" // Base URL for WHOOP API calls. Base URL + /v2/...

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "whoop-mcp", "config.json"), nil
}

// LoadConfig loads credentials and tokens
func LoadConfig() (*Config, error) {
	// Try loading from .env if it exists in the current directory
	_ = godotenv.Load() // ignore error, env vars might be set globally or in .env

	cfg := &Config{
		ClientID:     os.Getenv("WHOOP_CLIENT_ID"),
		ClientSecret: os.Getenv("WHOOP_CLIENT_SECRET"),
	}

	// Load tokens from config file if they exist
	configPath, err := GetConfigPath()
	if err != nil {
		return cfg, nil
	}

	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			var fileCfg Config
			if err := json.Unmarshal(data, &fileCfg); err == nil {
				// CLI/env vars take precedence for ClientID and ClientSecret
				if cfg.ClientID == "" {
					cfg.ClientID = fileCfg.ClientID
				}
				if cfg.ClientSecret == "" {
					cfg.ClientSecret = fileCfg.ClientSecret
				}
				cfg.AccessToken = fileCfg.AccessToken
				cfg.RefreshToken = fileCfg.RefreshToken
				cfg.ExpiresAt = fileCfg.ExpiresAt
			}
		}
	}

	return cfg, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(cfg *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}
