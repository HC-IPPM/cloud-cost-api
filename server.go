package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/bigquery"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/HC-IPPM/cloud-cost-api/auth"
	"github.com/HC-IPPM/cloud-cost-api/graph"
	"github.com/HC-IPPM/cloud-cost-api/middleware"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

const defaultPort = "8080"

func handleGenerateToken(w http.ResponseWriter, r *http.Request) {
	html := `<html><h1><a href="/login">Generate Access Token</a></h1></html>`
	fmt.Fprint(w, html)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	url := auth.GoogleOAuthConfig.AuthCodeURL("random-state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(w, r, url, http.StatusFound)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	token, err := auth.ExchangeCode(r.Context(), code)
	if err != nil {
		http.Error(w, "failed to exchange token", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Access Token: %s", token.AccessToken)
}

func main() {
	ctx := context.Background()
	projectID := `pdcp-serv-001-budgets`
	client, err := bigquery.NewClient(ctx, projectID, option.WithCredentialsFile("./service_account.json"))
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

	http.Handle("/", playground.Handler("GCP Cloud Cost Graohql playground", "/query"))
	http.Handle("/query", middleware.OAuth2Middleware(srv))
	http.HandleFunc("/token", handleGenerateToken)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/oauth2callback", handleCallback)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
