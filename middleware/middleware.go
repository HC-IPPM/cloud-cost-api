package middleware

import (
	"context"
	"fmt"
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
func OAuth2Middleware(ctx context.Context, success interface{}, failure interface{}) http.Handler {

	store := ctx.Value("sessionStore").(*persistence.Session)
	oauth := ctx.Value("oauth").(*auth.GcpOAuth)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token, sessionExists := store.GetSession(w, r)

		if authHeader == "" {
			if !sessionExists {
				// http.Error(w, "Missing Auth token", http.StatusUnauthorized)
				val, ok := failure.(string)
				if ok {
					http.Redirect(w, r, val, http.StatusFound)
					return
				} else {
					handler := failure.(http.Handler)
					handler.ServeHTTP(w, r)
				}
			}
		} else {
			token = strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				val, ok := failure.(string)
				if ok {
					http.Redirect(w, r, val, http.StatusFound)
					return
				} else {
					handler := failure.(http.Handler)
					handler.ServeHTTP(w, r)
				}
			}
		}

		if sessionExists {
			// Validate token with Google OAuth2
			userInfo, err := oauth.ValidateToken(token)
			if err != nil {
				fmt.Print("invalid or expired token ")
				val, ok := failure.(string)
				if ok {
					http.Redirect(w, r, val, http.StatusFound)
					return
				} else {
					handler := failure.(http.Handler)
					handler.ServeHTTP(w, r)
				}
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), userKey, userInfo)
			val, ok := success.(string)
			if ok {
				http.Redirect(w, r, val, http.StatusFound)
				return
			} else {
				handler := success.(http.Handler)
				handler.ServeHTTP(w, r.WithContext(ctx))
			}
		}
	})
}

func RateLimitMiddleware(ctx context.Context, next http.Handler) http.Handler {
	return nil
}

// FromContext retrieves user info from context
func FromContext(ctx context.Context) (map[string]interface{}, bool) {
	user, ok := ctx.Value(userKey).(map[string]interface{})
	return user, ok
}
