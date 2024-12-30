package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"net/http"

	"encoding/base64"
	"encoding/json"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Oauth interface {
	ExchangeCode(context.Context, string) (*oauth2.Token, error)
	AuthCodeURL(string, ...interface{}) (string)
	ValidateToken(string) (map[string]interface{}, error)
	GenerateState() string
	ValidateState(string) bool
}

type GcpOAuth struct {
	Config *oauth2.Config
	StateStore map[string]bool
}

func NewGoogleOAuth(ctx context.Context) *GcpOAuth {
	var GoogleOAuthConfig = &oauth2.Config{
		ClientID:     ctx.Value("oAuthClientID").(string),
		ClientSecret: ctx.Value("oAuthClientSecret").(string),
		RedirectURL:  "http://localhost:8080/oauth2callback",
		Scopes:       []string{"email", "profile", "openid"},
		Endpoint:     google.Endpoint,
	}

	gcp := new(GcpOAuth)
	gcp.Config = GoogleOAuthConfig
	gcp.StateStore = make(map[string]bool)
	return gcp
}

// ExchangeCode exchanges an authorization code for an access token.
func (gcp *GcpOAuth) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return gcp.Config.Exchange(ctx, code)
}

func (gcp *GcpOAuth) AuthCodeURL(state string, options ...interface{}) string {
	opts := []oauth2.AuthCodeOption{}
	for _, option := range options {
		opts = append(opts, option.(oauth2.AuthCodeOption))
	}
	return gcp.Config.AuthCodeURL(state, opts...)
}

func (gcp *GcpOAuth) ValidateToken(token string) (map[string]interface{}, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v3/userinfo?access_token=" + token)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, errors.New("Invalid or expired token")
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (gcp *GcpOAuth) GenerateState() string {
	b := make([]byte, 16)
	rand.Read(b) // Generate random bytes
	state := base64.URLEncoding.EncodeToString(b)
	gcp.StateStore[state] = true // Store state for later verification
	return state
}

func (gcp *GcpOAuth) ValidateState(state string) bool {
	return gcp.StateStore[state]
}
