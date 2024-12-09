package graph

import "cloud.google.com/go/bigquery"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	BigQueryClient *bigquery.Client
}

// NewResolver creates a new Resolver instance
func NewResolver(client *bigquery.Client) *Resolver {
	return &Resolver{
		BigQueryClient: client,
	}
}
