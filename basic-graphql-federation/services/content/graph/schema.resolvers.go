package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/marcelbeumer/go-playground/basic-graphql-federation/services/content/graph/generated"
	"github.com/marcelbeumer/go-playground/basic-graphql-federation/services/content/graph/model"
)

// TopProducts is the resolver for the topProducts field.
func (r *queryResolver) TopProducts(ctx context.Context, limit *int) (*model.Content, error) {
	title := "Title..."
	description := "Description..."

	var size int
	if limit != nil {
		size = *limit
	} else {
		size = 10
	}

	return &model.Content{
		Title:       &title,
		Description: &description,
		Products:    r.Resolver.TopProducts[:size],
	}, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
