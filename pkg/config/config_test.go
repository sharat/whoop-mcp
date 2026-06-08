package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigUsesEnvironmentBeforeDotEnvAndFile(t *testing.T) {
	home := prepareConfigTest(t)
	t.Setenv("WHOOP_CLIENT_ID", "environment-id")
	t.Setenv("WHOOP_CLIENT_SECRET", "environment-secret")

	writeDotEnv(t, "dotenv-id", "dotenv-secret")
	expiresAt := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	writeConfigFile(t, home, Config{
		ClientID:     "file-id",
		ClientSecret: "file-secret",
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    expiresAt,
	})

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.ClientID != "environment-id" {
		t.Errorf("ClientID = %q, want environment-id", cfg.ClientID)
	}
	if cfg.ClientSecret != "environment-secret" {
		t.Errorf("ClientSecret = %q, want environment-secret", cfg.ClientSecret)
	}
	if cfg.AccessToken != "access-token" || cfg.RefreshToken != "refresh-token" {
		t.Errorf("tokens = %q, %q, want values from config file", cfg.AccessToken, cfg.RefreshToken)
	}
	if !cfg.ExpiresAt.Equal(expiresAt) {
		t.Errorf("ExpiresAt = %v, want %v", cfg.ExpiresAt, expiresAt)
	}
}

func TestLoadConfigUsesDotEnvBeforeFile(t *testing.T) {
	home := prepareConfigTest(t)
	writeDotEnv(t, "dotenv-id", "dotenv-secret")
	writeConfigFile(t, home, Config{
		ClientID:     "file-id",
		ClientSecret: "file-secret",
	})

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.ClientID != "dotenv-id" {
		t.Errorf("ClientID = %q, want dotenv-id", cfg.ClientID)
	}
	if cfg.ClientSecret != "dotenv-secret" {
		t.Errorf("ClientSecret = %q, want dotenv-secret", cfg.ClientSecret)
	}
}

func TestLoadConfigFallsBackToFileCredentials(t *testing.T) {
	home := prepareConfigTest(t)
	writeConfigFile(t, home, Config{
		ClientID:     "file-id",
		ClientSecret: "file-secret",
	})

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.ClientID != "file-id" {
		t.Errorf("ClientID = %q, want file-id", cfg.ClientID)
	}
	if cfg.ClientSecret != "file-secret" {
		t.Errorf("ClientSecret = %q, want file-secret", cfg.ClientSecret)
	}
}

func TestLoadConfigUsesDefaultCallbackPort(t *testing.T) {
	prepareConfigTest(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.CallbackPort != DefaultCallbackPort {
		t.Errorf("CallbackPort = %d, want %d", cfg.CallbackPort, DefaultCallbackPort)
	}
	if cfg.RedirectURI() != "http://127.0.0.1:8181/callback" {
		t.Errorf("RedirectURI() = %q", cfg.RedirectURI())
	}
}

func TestLoadConfigUsesCallbackPortFromEnvironment(t *testing.T) {
	prepareConfigTest(t)
	t.Setenv("WHOOP_CALLBACK_PORT", "9191")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.CallbackPort != 9191 {
		t.Errorf("CallbackPort = %d, want 9191", cfg.CallbackPort)
	}
	if cfg.CallbackAddress() != "127.0.0.1:9191" {
		t.Errorf("CallbackAddress() = %q", cfg.CallbackAddress())
	}
	if cfg.RedirectURI() != "http://127.0.0.1:9191/callback" {
		t.Errorf("RedirectURI() = %q", cfg.RedirectURI())
	}
}

func TestLoadConfigRejectsInvalidCallbackPort(t *testing.T) {
	tests := []string{"not-a-port", "0", "65536"}
	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			prepareConfigTest(t)
			t.Setenv("WHOOP_CALLBACK_PORT", value)

			if _, err := LoadConfig(); err == nil {
				t.Fatalf("LoadConfig() error = nil, want invalid port error")
			}
		})
	}
}

func prepareConfigTest(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	workDir := filepath.Join(dir, "work")
	if err := os.MkdirAll(home, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workDir, 0700); err != nil {
		t.Fatal(err)
	}

	oldWorkDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWorkDir); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	t.Setenv("HOME", home)
	unsetEnv(t, "WHOOP_CLIENT_ID")
	unsetEnv(t, "WHOOP_CLIENT_SECRET")
	unsetEnv(t, "WHOOP_CALLBACK_PORT")
	return home
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	value, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		var err error
		if existed {
			err = os.Setenv(key, value)
		} else {
			err = os.Unsetenv(key)
		}
		if err != nil {
			t.Errorf("restore %s: %v", key, err)
		}
	})
}

func writeDotEnv(t *testing.T, clientID, clientSecret string) {
	t.Helper()
	data := []byte("WHOOP_CLIENT_ID=" + clientID + "\nWHOOP_CLIENT_SECRET=" + clientSecret + "\n")
	if err := os.WriteFile(".env", data, 0600); err != nil {
		t.Fatal(err)
	}
}

func writeConfigFile(t *testing.T, home string, cfg Config) {
	t.Helper()
	path := filepath.Join(home, ".config", "whoop-mcp", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}
