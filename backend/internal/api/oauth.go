package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig *oauth2.Config
	oauthStateString  = "random-state-string" // In production, use a secure random state
)

func initOAuthConfig() {
	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"), // e.g., https://status.etsusa.com/api/v1/auth/google/callback
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func (s *Server) handleGoogleLogin(c *gin.Context) {
	if googleOauthConfig == nil {
		initOAuthConfig()
	}

	if googleOauthConfig.ClientID == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google OAuth not configured"})
		return
	}

	url := googleOauthConfig.AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (s *Server) handleGoogleCallback(c *gin.Context) {
	if googleOauthConfig == nil {
		initOAuthConfig()
	}

	state := c.Query("state")
	if state != oauthStateString {
		fmt.Printf("OAuth callback error: Invalid state parameter\n")
		c.Redirect(http.StatusTemporaryRedirect, "/?error=invalid_state")
		return
	}

	code := c.Query("code")
	if code == "" {
		fmt.Printf("OAuth callback error: Code not found\n")
		c.Redirect(http.StatusTemporaryRedirect, "/?error=no_code")
		return
	}

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("OAuth callback error: Failed to exchange token: %v\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=token_exchange_failed")
		return
	}

	// Get user info from Google
	userInfo, err := getUserInfoFromGoogle(token.AccessToken)
	if err != nil {
		fmt.Printf("OAuth callback error: Failed to get user info: %v\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=userinfo_failed")
		return
	}

	fmt.Printf("OAuth: Got user info for %s (%s)\n", userInfo.Email, userInfo.Name)

	// Check if email domain is etsusa.com
	if !strings.HasSuffix(userInfo.Email, "@etsusa.com") {
		fmt.Printf("OAuth: Unauthorized domain for email: %s\n", userInfo.Email)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=unauthorized_domain")
		return
	}

	// Check if user exists, if not create them
	user, err := s.postgres.GetUserByUsername(context.Background(), userInfo.Email)
	if err != nil {
		// User doesn't exist, create them
		fmt.Printf("OAuth: Creating new user for %s\n", userInfo.Email)
		user, err = s.postgres.CreateUserFromOAuth(context.Background(), userInfo.Email, userInfo.Name)
		if err != nil {
			fmt.Printf("OAuth callback error: Failed to create user: %v\n", err)
			c.Redirect(http.StatusTemporaryRedirect, "/?error=user_creation_failed")
			return
		}
	}

	// Generate JWT token
	jwtToken, err := generateToken(user)
	if err != nil {
		fmt.Printf("OAuth callback error: Failed to generate token: %v\n", err)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=token_generation_failed")
		return
	}

	fmt.Printf("OAuth: Successfully authenticated user %s, redirecting with token\n", userInfo.Email)

	// Redirect to login page with token - login page will handle auth setup
	host := c.Request.Host
	redirectURL := fmt.Sprintf("https://%s/login?token=%s", host, jwtToken)

	fmt.Printf("OAuth: Redirecting to: %s\n", redirectURL)
	// Redirect to frontend with token
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func getUserInfoFromGoogle(accessToken string) (*GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
