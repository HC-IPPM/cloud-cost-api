package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	// "github.com/HC-IPPM/cloud-cost-api/auth"
)

type contextKey string

const userKey contextKey = "user"

// Middleware to validate OAuth2 token
func OAuth2Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
			return
		}

		// Validate token with Google OAuth2
		userInfo, err := validateToken(token)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), userKey, userInfo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken fetches user info to validate the token
func validateToken(token string) (map[string]interface{}, error) {
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

// FromContext retrieves user info from context
func FromContext(ctx context.Context) (map[string]interface{}, bool) {
	user, ok := ctx.Value(userKey).(map[string]interface{})
	return user, ok
}