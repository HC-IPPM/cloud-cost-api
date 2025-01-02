package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/HC-IPPM/cloud-cost-api/auth"
	"github.com/HC-IPPM/cloud-cost-api/graph"
	"github.com/HC-IPPM/cloud-cost-api/middleware"
	"github.com/HC-IPPM/cloud-cost-api/persistence"
	_ "github.com/googleapis/enterprise-certificate-proxy/client"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

type contextKey string

var SessionStore *persistence.Session
var Oauth auth.Oauth

type Secrets struct {
	OAuthClientID     string `json:"oAuthClientID"`
	OAuthClientSecret string `json:"oAuthClientSecret"`
	OAuthCallbackURL  string `json:"oAuthCallbackURL"`
	SessionKey        string `json:"sessionKey"`
}

const defaultPort = "8080"

func SetAccessTokenCookie(w http.ResponseWriter, accessToken string) {
	cookie := &http.Cookie{
		Name:     "access_token",                // Name of the cookie
		Value:    accessToken,                   // Value of the cookie (access token)
		Path:     "/",                           // Scope of the cookie
		Expires:  time.Now().Add(1 * time.Hour), // Set expiration (adjust based on your token expiry)
		HttpOnly: true,                          // Ensure the cookie is not accessible via JavaScript
		Secure:   true,                          // Ensure the cookie is only sent over HTTPS
		SameSite: http.SameSiteStrictMode,       // SameSite policy to protect against CSRF
	}

	http.SetCookie(w, cookie)
}

func handleGenerateToken(w http.ResponseWriter, r *http.Request) {
	html := `<html><h1><a href="/login">Login With GCP</a></h1></html>`
	fmt.Fprint(w, html)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	randomState := Oauth.GenerateState()
	url := Oauth.AuthCodeURL(randomState, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(w, r, url, http.StatusFound)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	SessionStore.ClearSession(w, r)
	html := `<html><h1>Logged Out</h1></html>`
	fmt.Fprintf(w, html)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if !Oauth.ValidateState(state) {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	token, err := Oauth.ExchangeCode(r.Context(), code)
	if err != nil {
		http.Error(w, "failed to exchange token", http.StatusInternalServerError)
		return
	}

	SessionStore.SetSession(w, r, token.AccessToken)
	http.Redirect(w, r, "/", http.StatusFound)
}

func loadSecrets(ctx context.Context, filePath string) context.Context {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}
	defer file.Close()

	// Create a decoder and parse the JSON
	var secrets Secrets
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&secrets)
	if err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		return nil
	}

	ctx = context.WithValue(ctx, "sessionKey", secrets.SessionKey)
	ctx = context.WithValue(ctx, "oAuthClientID", secrets.OAuthClientID)
	ctx = context.WithValue(ctx, "oAuthClientSecret", secrets.OAuthClientSecret)
	ctx = context.WithValue(ctx, "oAuthCallbackURL", secrets.OAuthCallbackURL)
	return ctx
}

func BigQueryClient(ctx context.Context, projectID string) (*bigquery.Client, error) {
	// Application Default Credentials will be used if the service account key file 
	// does not exist
	_, err := os.Stat("./service_account.json")
	if os.IsNotExist(err) {
		return bigquery.NewClient(ctx, projectID)
	} else {
		return bigquery.NewClient(ctx, projectID, option.WithCredentialsFile("./service_account.json"))
	}
}

func main() {
	ctx := context.Background()

	ctx = loadSecrets(ctx, "/tmp/secrets.json")

	SessionStore = persistence.NewCookieStore(ctx)
	ctx = context.WithValue(ctx, "sessionStore", SessionStore)

	Oauth = auth.NewGoogleOAuth(ctx)
	ctx = context.WithValue(ctx, "oauth", Oauth)

	projectID := `pdcp-serv-001-budgets`
	client, err := BigQueryClient(ctx, projectID)

	if err != nil {
		fmt.Errorf(err.Error())
	}
	defer client.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	resolver := graph.NewResolver(client)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.Handle("/", middleware.OAuth2Middleware(
		ctx, playground.Handler("GCP Cloud Cost Graphql playground", "/query")))
	http.Handle("/query", middleware.OAuth2Middleware(ctx, srv))
	http.HandleFunc("/token", handleGenerateToken)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/logout", handleLogout)
	http.HandleFunc("/oauth2callback", handleCallback)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
