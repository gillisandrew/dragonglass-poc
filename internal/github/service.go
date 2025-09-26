// ABOUTME: GitHub API adapter - wraps GitHub authentication and API operations
// ABOUTME: Implements domain.AuthService interface using GitHub OAuth device flow
package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pterm/pterm"
	"github.com/zalando/go-keyring"
)

// Service implements domain.AuthService using GitHub OAuth
type Service struct {
	gitHubHost     string
	requiredScopes string
	httpClient     *http.Client
}

// NewService creates a new GitHub authentication service
func NewService() *Service {
	return &Service{
		gitHubHost:     "github.com",
		requiredScopes: "read:packages",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Authenticate implements domain.AuthService.Authenticate
func (s *Service) Authenticate() error {
	return s.authenticateWithDeviceFlow()
}

// IsAuthenticated implements domain.AuthService.IsAuthenticated
func (s *Service) IsAuthenticated() bool {
	token, err := s.GetToken()
	if err != nil || token == "" {
		return false
	}

	// Validate the token by making a test request
	return s.validateToken(token) == nil
}

// GetToken implements domain.AuthService.GetToken
func (s *Service) GetToken() (string, error) {
	cred, err := s.getStoredCredential()
	if err != nil {
		return "", fmt.Errorf("no stored credentials found: %w", err)
	}
	return cred.Token, nil
}

// GetUser implements domain.AuthService.GetUser
func (s *Service) GetUser() (string, error) {
	cred, err := s.getStoredCredential()
	if err != nil {
		return "", fmt.Errorf("not authenticated: %w", err)
	}

	if cred.Username != "" {
		return cred.Username, nil
	}

	// Fetch user info from GitHub API
	token := cred.Token
	return s.fetchAuthenticatedUser(token)
}

// Logout implements domain.AuthService.Logout
func (s *Service) Logout() error {
	return s.clearStoredToken()
}

// Internal types and methods

type storedCredential struct {
	Token     string    `json:"token"`
	Scopes    string    `json:"scopes"`
	Username  string    `json:"username,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Source    string    `json:"source"`
}

type deviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
}

// Constants for keychain storage
const (
	keyringService = "dragonglass-cli"
	keyringAccount = "github-token"
)

// getStoredCredential retrieves stored credentials from keychain or file fallback
func (s *Service) getStoredCredential() (*storedCredential, error) {
	// Try keychain first
	data, err := keyring.Get(keyringService, keyringAccount)
	if err == nil {
		var cred storedCredential
		if err := json.Unmarshal([]byte(data), &cred); err != nil {
			return nil, fmt.Errorf("failed to unmarshal credential: %w", err)
		}
		return &cred, nil
	}

	// TODO: Add file fallback implementation
	return nil, fmt.Errorf("failed to get credentials from keychain: %w", err)
}

// clearStoredToken removes stored credentials
func (s *Service) clearStoredToken() error {
	_ = keyring.Delete(keyringService, keyringAccount)
	// TODO: Also clear file fallback
	return nil
}

// validateToken checks if the token is valid by making a test request
func (s *Service) validateToken(token string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.%s/user", s.gitHubHost), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token validation failed with status: %d", resp.StatusCode)
	}

	return nil
}

// fetchAuthenticatedUser gets user information from GitHub API
func (s *Service) fetchAuthenticatedUser(token string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.%s/user", s.gitHubHost), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch user, status: %d", resp.StatusCode)
	}

	var user struct {
		Login string `json:"login"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("failed to decode user response: %w", err)
	}

	return user.Login, nil
}

// authenticateWithDeviceFlow implements GitHub OAuth device flow
func (s *Service) authenticateWithDeviceFlow() error {
	pterm.Info.Println("Starting GitHub device flow authentication")

	// Step 1: Request device code
	deviceCode, err := s.requestDeviceCode()
	if err != nil {
		return fmt.Errorf("failed to request device code: %w", err)
	}

	// Step 2: Display user code and verification URL
	pterm.Info.Printf("Please visit: %s\n", deviceCode.VerificationURI)
	pterm.Info.Printf("And enter code: %s\n", deviceCode.UserCode)

	// Step 3: Poll for access token
	token, err := s.pollForAccessToken(deviceCode)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Step 4: Get user information
	username, err := s.fetchAuthenticatedUser(token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	// Step 5: Store credentials
	cred := storedCredential{
		Token:     token.AccessToken,
		Scopes:    token.Scope,
		Username:  username,
		CreatedAt: time.Now(),
		Source:    "device-flow",
	}

	return s.storeCredential(cred)
}

// requestDeviceCode requests a device code from GitHub
func (s *Service) requestDeviceCode() (*deviceCodeResponse, error) {
	// Implementation would go here - simplified for now
	return nil, fmt.Errorf("device flow not implemented yet")
}

// pollForAccessToken polls GitHub for the access token
func (s *Service) pollForAccessToken(deviceCode *deviceCodeResponse) (*accessTokenResponse, error) {
	// Implementation would go here - simplified for now
	return nil, fmt.Errorf("token polling not implemented yet")
}

// storeCredential saves credentials to keychain
func (s *Service) storeCredential(cred storedCredential) error {
	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("failed to marshal credential: %w", err)
	}

	if err := keyring.Set(keyringService, keyringAccount, string(data)); err != nil {
		return fmt.Errorf("failed to store credential: %w", err)
	}

	pterm.Success.Println("🔐 Token stored securely in OS keychain")
	return nil
}
