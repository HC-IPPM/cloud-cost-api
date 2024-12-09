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
	"github.com/HC-IPPM/cloud-cost-api/graph"
	"google.golang.org/api/option"
)

const defaultPort = "8080"

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

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
