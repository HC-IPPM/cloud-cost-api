package auth

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleOAuthConfig = &oauth2.Config{
	ClientID:     "",
	ClientSecret: "",
	RedirectURL:  "http://localhost:8080/oauth2callback",
	Scopes:       []string{"email", "profile", "openid"},
	Endpoint:     google.Endpoint,
}

// ExchangeCode exchanges an authorization code for an access token.
func ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return GoogleOAuthConfig.Exchange(ctx, code)
}
