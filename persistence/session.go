package persistence

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
)

type contextKey string

type Session struct {
	SessionKey string
	Store      sessions.Store
}

func NewCookieStore(ctx context.Context) *Session {
	sesh := new(Session)
	sesh.SessionKey = ctx.Value("sessionKey").(string)
	cookieStore := sessions.NewCookieStore([]byte(sesh.SessionKey))
	sesh.Store = cookieStore
	return sesh
}

func (sesh *Session) SetSession(w http.ResponseWriter, r *http.Request, accessToken string) {
	session, _ := sesh.Store.Get(r, "user-session")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	session.Values["access_token"] = accessToken
	session.Save(r, w)
}

func (sesh *Session) GetSession(w http.ResponseWriter, r *http.Request) (string, bool) {
	// Get the session
	session, _ := sesh.Store.Get(r, "user-session")
	token, ok := session.Values["access_token"].(string)
	return token, ok
}

func (sesh *Session) ClearSession(w http.ResponseWriter, r *http.Request) {
	// Get the session
	session, _ := sesh.Store.Get(r, "user-session")

	// Invalidate the session
	session.Options.MaxAge = -1
	session.Save(r, w)
}

func GetAccessTokenCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
