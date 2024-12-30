package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/HC-IPPM/cloud-cost-api/auth"
	"github.com/HC-IPPM/cloud-cost-api/persistence"
)

type contextKey string

const userKey contextKey = "user"

func GetAccessTokenCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// Middleware to validate OAuth2 token
func OAuth2Middleware(ctx context.Context, next http.Handler) http.Handler {

	store := ctx.Value("sessionStore").(*persistence.Session)
	oauth := ctx.Value("oauth").(*auth.GcpOAuth)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token, sessionExists := store.GetSession(w, r)
		if authHeader == "" {
			if !sessionExists {
				// http.Error(w, "Missing Auth token", http.StatusUnauthorized)
				http.Redirect(w, r, "/token", http.StatusFound)
				return
			}
		} else {
			token = strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				// http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
				http.Redirect(w, r, "/token", http.StatusFound)
				return
			}
		}

		// Validate token with Google OAuth2
		userInfo, err := oauth.ValidateToken(token)
		if err != nil {
			// http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			http.Redirect(w, r, "/token", http.StatusFound)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), userKey, userInfo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FromContext retrieves user info from context
func FromContext(ctx context.Context) (map[string]interface{}, bool) {
	user, ok := ctx.Value(userKey).(map[string]interface{})
	return user, ok
}
