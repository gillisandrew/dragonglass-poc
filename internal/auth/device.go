// ABOUTME: Native GitHub device flow implementation for package registry access
// ABOUTME: Implements OAuth device flow with specific scopes for ghcr.io access
package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

const (
	// GitHub App configuration for dragonglass
	// Note: These would typically be from environment or config in production
	ClientID       = "Iv23liqsBDwGGj4wPGbZ" // GitHub App client ID for dragonglass
	DeviceCodeURL  = "https://github.com/login/device/code"
	AccessTokenURL = "https://github.com/login/oauth/access_token"

	// Scopes required for ghcr.io access
	PackageReadScope = "read:packages"
	DefaultScopes    = PackageReadScope
)

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type AccessTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// StartDeviceFlow initiates the GitHub device flow authentication
func StartDeviceFlow(scopes string) (*DeviceCodeResponse, error) {
	if scopes == "" {
		scopes = DefaultScopes
	}

	// Prepare form data
	data := url.Values{}
	data.Set("client_id", ClientID)
	data.Set("scope", scopes)

	// Make request to GitHub
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(DeviceCodeURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // Ignore error on close
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub device code request failed: %d - %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse URL-encoded response
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse form response: %w", err)
	}

	// Convert expires_in to int
	expiresIn, err := strconv.Atoi(values.Get("expires_in"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse expires_in: %w", err)
	}

	// Convert interval to int
	interval, err := strconv.Atoi(values.Get("interval"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse interval: %w", err)
	}

	// Build response struct
	deviceResp := DeviceCodeResponse{
		DeviceCode:      values.Get("device_code"),
		UserCode:        values.Get("user_code"),
		VerificationURI: values.Get("verification_uri"),
		ExpiresIn:       expiresIn,
		Interval:        interval,
	}

	return &deviceResp, nil
}

// PollForAccessToken polls GitHub for the access token
func PollForAccessToken(ctx context.Context, deviceCode string, interval int) (*AccessTokenResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("authentication timed out")
		case <-ticker.C:
			// Prepare form data
			data := url.Values{}
			data.Set("client_id", ClientID)
			data.Set("device_code", deviceCode)
			data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

			// Make token request
			resp, err := client.PostForm(AccessTokenURL, data)
			if err != nil {
				return nil, fmt.Errorf("failed to request access token: %w", err)
			}

			body, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close() // Ignore error on close

			if err != nil {
				return nil, fmt.Errorf("failed to read token response: %w", err)
			}

			// Parse URL-encoded token response
			values, err := url.ParseQuery(string(body))
			if err != nil {
				return nil, fmt.Errorf("failed to parse token response: %w", err)
			}

			// Build token response struct
			tokenResp := AccessTokenResponse{
				AccessToken:      values.Get("access_token"),
				TokenType:        values.Get("token_type"),
				Scope:            values.Get("scope"),
				Error:            values.Get("error"),
				ErrorDescription: values.Get("error_description"),
			}

			// Handle different response states
			switch tokenResp.Error {
			case "":
				// Success!
				return &tokenResp, nil
			case "authorization_pending":
				// User hasn't completed the flow yet, continue polling
				continue
			case "slow_down":
				// GitHub wants us to slow down, double the interval
				ticker.Reset(time.Duration(interval*2) * time.Second)
				continue
			case "expired_token":
				return nil, fmt.Errorf("device code expired, please try again")
			case "access_denied":
				return nil, fmt.Errorf("access denied by user")
			default:
				return nil, fmt.Errorf("GitHub OAuth error: %s - %s", tokenResp.Error, tokenResp.ErrorDescription)
			}
		}
	}
}

// RunDeviceFlow executes the complete device flow authentication
func RunDeviceFlow(scopes string) (*AccessTokenResponse, error) {
	pterm.Info.Println("Starting GitHub device flow authentication")
	pterm.Info.Printfln("Requesting scopes: %s", pterm.LightMagenta(scopes))
	pterm.Println()

	// Step 1: Get device code
	deviceCode, err := StartDeviceFlow(scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to start device flow: %w", err)
	}

	// Step 2: Show user instructions with pterm
	pterm.Info.Println("Please complete the following steps:")
	pterm.Println()

	// Create a styled box for the user code
	codeBox := pterm.DefaultBox.WithTitle("Code").WithTitleTopCenter()
	codeBox.Println(pterm.LightCyan(deviceCode.UserCode))
	pterm.Println()

	// Show instructions as a list
	instructions := pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
		{Level: 0, Text: pterm.Sprintf("Copy the code above: %s", pterm.LightCyan(deviceCode.UserCode))},
		{Level: 0, Text: pterm.Sprintf("Visit: %s", pterm.LightBlue(deviceCode.VerificationURI))},
		{Level: 0, Text: "Enter the code when prompted"},
	})
	instructions.Render()
	pterm.Println()

	// Show expiration and polling status
	pterm.Warning.Printfln("Code expires in %d minutes", deviceCode.ExpiresIn/60)
	spinner, _ := pterm.DefaultSpinner.Start("Polling for authorization...")

	// Step 3: Poll for access token
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(deviceCode.ExpiresIn)*time.Second)
	defer cancel()

	token, err := PollForAccessToken(ctx, deviceCode.DeviceCode, deviceCode.Interval)
	spinner.Stop()
	if err != nil {
		pterm.Error.Println("Failed to get access token")
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	pterm.Success.Println("Authentication successful!")
	pterm.Info.Printfln("Received token with scopes: %s", pterm.LightGreen(token.Scope))

	return token, nil
}

// ValidateTokenScopes checks if the token has the required scopes
func ValidateTokenScopes(token string, requiredScopes []string) error {
	// Create request to get token info
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("User-Agent", "dragonglass-cli")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // Ignore error on close
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token validation failed: %d", resp.StatusCode)
	}

	// Check X-OAuth-Scopes header for granted scopes
	grantedScopes := resp.Header.Get("X-OAuth-Scopes")
	if grantedScopes == "" {
		return fmt.Errorf("unable to determine token scopes")
	}

	// Parse granted scopes
	scopes := strings.Split(grantedScopes, ",")
	scopeSet := make(map[string]bool)
	for _, scope := range scopes {
		scopeSet[strings.TrimSpace(scope)] = true
	}

	// Check if all required scopes are present
	for _, required := range requiredScopes {
		if !scopeSet[required] {
			return fmt.Errorf("missing required scope: %s (granted: %s)", required, grantedScopes)
		}
	}

	return nil
}
