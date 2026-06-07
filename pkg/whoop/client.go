package whoop

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"whoop-mcp/pkg/config"
)

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Config returns the internal configuration (useful for debugging/verifying tokens)
func (c *Client) Config() *config.Config {
	return c.cfg
}

// ensureValidToken checks if the token is close to expiry and refreshes if needed
func (c *Client) ensureValidToken(ctx context.Context) error {
	if c.cfg.AccessToken == "" {
		return fmt.Errorf("no access token available, please authenticate first")
	}
	// Refresh if expired or expiring in less than 5 minutes
	if c.cfg.ExpiresAt.IsZero() || time.Now().Add(5*time.Minute).After(c.cfg.ExpiresAt) {
		return c.RefreshToken(ctx)
	}
	return nil
}

// RefreshToken refreshes the OAuth 2.0 access token using the rotating refresh token
func (c *Client) RefreshToken(ctx context.Context) error {
	if c.cfg.ClientID == "" || c.cfg.ClientSecret == "" {
		return fmt.Errorf("missing client credentials (WHOOP_CLIENT_ID or WHOOP_CLIENT_SECRET)")
	}
	if c.cfg.RefreshToken == "" {
		return fmt.Errorf("no refresh token available, please login first")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", c.cfg.RefreshToken)
	data.Set("client_id", c.cfg.ClientID)
	data.Set("client_secret", c.cfg.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return err
	}

	c.cfg.AccessToken = tokenResp.AccessToken
	c.cfg.RefreshToken = tokenResp.RefreshToken
	c.cfg.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Save the updated rotating tokens to configuration file
	return config.SaveConfig(c.cfg)
}

// doRequest performs an authenticated HTTP GET request to WHOOP API v2
func (c *Client) doRequest(ctx context.Context, path string, queryParams map[string]string) ([]byte, error) {
	if err := c.ensureValidToken(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	u, err := url.Parse(config.BaseURL + path)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	for k, v := range queryParams {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("WHOOP API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetProfile retrieves the user's basic profile
func (c *Client) GetProfile(ctx context.Context) ([]byte, error) {
	return c.doRequest(ctx, "/v2/user/profile/basic", nil)
}

// GetBodyMeasurements retrieves the user's body measurements
func (c *Client) GetBodyMeasurements(ctx context.Context) ([]byte, error) {
	return c.doRequest(ctx, "/v2/user/measurement/body", nil)
}

// GetCycles retrieves a collection of physiological cycles
func (c *Client) GetCycles(ctx context.Context, start, end, limit, nextToken string) ([]byte, error) {
	params := map[string]string{
		"start":      start,
		"end":        end,
		"limit":      limit,
		"nextToken":  nextToken,
	}
	return c.doRequest(ctx, "/v2/cycle", params)
}

// GetSleeps retrieves a collection of sleep activities
func (c *Client) GetSleeps(ctx context.Context, start, end, limit, nextToken string) ([]byte, error) {
	params := map[string]string{
		"start":      start,
		"end":        end,
		"limit":      limit,
		"nextToken":  nextToken,
	}
	return c.doRequest(ctx, "/v2/activity/sleep", params)
}

// GetRecoveries retrieves a collection of recovery metrics
func (c *Client) GetRecoveries(ctx context.Context, start, end, limit, nextToken string) ([]byte, error) {
	params := map[string]string{
		"start":      start,
		"end":        end,
		"limit":      limit,
		"nextToken":  nextToken,
	}
	return c.doRequest(ctx, "/v2/recovery", params)
}

// GetWorkouts retrieves a collection of workout activities
func (c *Client) GetWorkouts(ctx context.Context, start, end, limit, nextToken string) ([]byte, error) {
	params := map[string]string{
		"start":      start,
		"end":        end,
		"limit":      limit,
		"nextToken":  nextToken,
	}
	return c.doRequest(ctx, "/v2/activity/workout", params)
}
