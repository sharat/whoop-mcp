package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/sharat/whoop-mcp/pkg/config"
	"github.com/sharat/whoop-mcp/pkg/mcp"
	"github.com/sharat/whoop-mcp/pkg/whoop"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	args := os.Args
	if len(args) > 1 && args[1] == "login" {
		handleLogin(cfg)
		return
	}

	// Default behavior is to start the MCP server
	handleRun(cfg)
}

func handleLogin(cfg *config.Config) {
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		fmt.Println("Error: WHOOP_CLIENT_ID and WHOOP_CLIENT_SECRET must be set in your .env file or environment.")
		fmt.Println("Please make sure they are defined in: /Users/sarat/github/whoop-mcp/.env")
		os.Exit(1)
	}

	state := fmt.Sprintf("%d", time.Now().UnixNano())
	
	// Scopes required for WHOOP API v2
	scopes := "offline read:profile read:body_measurement read:cycles read:sleep read:recovery read:workout"
	
	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		config.AuthURL,
		url.QueryEscape(cfg.ClientID),
		url.QueryEscape(config.RedirectURI),
		url.QueryEscape(scopes),
		state,
	)

	fmt.Println("=========================================================")
	fmt.Println("           WHOOP MCP Server Authorization Flow           ")
	fmt.Println("=========================================================")
	fmt.Println("1. Open the following URL in your browser to log in and authorize WHOOP access:")
	fmt.Println()
	fmt.Println(authURL)
	fmt.Println()
	fmt.Println("2. Once authorized, you will be automatically redirected.")
	fmt.Println("   Waiting for callback on http://127.0.0.1:8181/callback ...")
	fmt.Println("=========================================================")

	srv := &http.Server{Addr: ":8181"}
	var loginErr error

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		cbState := q.Get("state")
		code := q.Get("code")

		if cbState != state {
			msg := "Error: state mismatch. Security verification failed."
			http.Error(w, msg, http.StatusBadRequest)
			loginErr = fmt.Errorf(msg)
			go func() { _ = srv.Shutdown(context.Background()) }()
			return
		}

		if code == "" {
			msg := "Error: authorization code not provided in callback."
			http.Error(w, msg, http.StatusBadRequest)
			loginErr = fmt.Errorf(msg)
			go func() { _ = srv.Shutdown(context.Background()) }()
			return
		}

		// Exchange code for token
		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Set("code", code)
		data.Set("redirect_uri", config.RedirectURI)
		data.Set("client_id", cfg.ClientID)
		data.Set("client_secret", cfg.ClientSecret)

		resp, err := http.PostForm(config.TokenURL, data)
		if err != nil {
			msg := fmt.Sprintf("Failed to request tokens: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			loginErr = fmt.Errorf(msg)
			go func() { _ = srv.Shutdown(context.Background()) }()
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			msg := fmt.Sprintf("Failed to read token response: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			loginErr = fmt.Errorf(msg)
			go func() { _ = srv.Shutdown(context.Background()) }()
			return
		}

		if resp.StatusCode != http.StatusOK {
			msg := fmt.Sprintf("Token exchange failed with status %d: %s", resp.StatusCode, string(body))
			http.Error(w, msg, http.StatusBadRequest)
			loginErr = fmt.Errorf(msg)
			go func() { _ = srv.Shutdown(context.Background()) }()
			return
		}

		var tokenResp struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
		}

		if err := json.Unmarshal(body, &tokenResp); err != nil {
			msg := fmt.Sprintf("Failed to parse token response: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			loginErr = fmt.Errorf(msg)
			go func() { _ = srv.Shutdown(context.Background()) }()
			return
		}

		cfg.AccessToken = tokenResp.AccessToken
		cfg.RefreshToken = tokenResp.RefreshToken
		cfg.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

		if err := config.SaveConfig(cfg); err != nil {
			msg := fmt.Sprintf("Failed to save configuration: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			loginErr = fmt.Errorf(msg)
			go func() { _ = srv.Shutdown(context.Background()) }()
			return
		}

		// Premium response HTML
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<meta charset="utf-8">
				<meta name="viewport" content="width=device-width, initial-scale=1">
				<title>WHOOP MCP Authorization Successful</title>
				<style>
					body {
						font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
						text-align: center;
						padding: 80px 20px;
						background-color: #0b0f19;
						color: #f3f4f6;
						margin: 0;
					}
					.card {
						background: #111827;
						padding: 40px;
						border-radius: 12px;
						box-shadow: 0 10px 15px -3px rgba(0,0,0,0.3);
						display: inline-block;
						max-width: 500px;
						border: 1px solid #1f2937;
					}
					h1 {
						color: #10b981;
						margin-top: 0;
						margin-bottom: 20px;
						font-size: 28px;
						font-weight: 700;
					}
					p {
						color: #9ca3af;
						font-size: 16px;
						line-height: 1.6;
					}
					code {
						background: #1f2937;
						color: #38bdf8;
						padding: 4px 8px;
						border-radius: 4px;
						font-family: monospace;
						font-size: 14px;
					}
				</style>
			</head>
			<body>
				<div class="card">
					<h1>Authentication Successful!</h1>
					<p>Your WHOOP API v2 tokens have been successfully generated and saved to:</p>
					<p><code>~/.config/whoop-mcp/config.json</code></p>
					<p>You can now close this browser tab and return to your terminal.</p>
				</div>
			</body>
			</html>
		`))

		fmt.Println("Authorization successful! Tokens saved.")
		go func() { _ = srv.Shutdown(context.Background()) }()
	})

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Printf("HTTP server error: %v\n", err)
		os.Exit(1)
	}

	if loginErr != nil {
		fmt.Printf("Login failed: %v\n", loginErr)
		os.Exit(1)
	}
}

func handleRun(cfg *config.Config) {
	if cfg.AccessToken == "" {
		fmt.Println("Error: No access token found.")
		fmt.Println("Please run the login flow first to authenticate:")
		fmt.Println("  go run cmd/whoop-mcp/main.go login")
		os.Exit(1)
	}

	client := whoop.NewClient(cfg)
	mcpSrv := mcp.NewServer(client)

	ctx := context.Background()
	if err := mcpSrv.Run(ctx); err != nil {
		log.Fatalf("MCP Server run error: %v", err)
	}
}
