package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/marcelbeumer/go-playground/basic-graphql-federation/services/commerce/graph/generated"
	"github.com/marcelbeumer/go-playground/basic-graphql-federation/services/commerce/graph/model"
)

// Product is the resolver for the product field.
func (r *queryResolver) Product(ctx context.Context, sku *string) (*model.Product, error) {
	for _, p := range r.TopProducts {
		if p.Sku == *sku {
			return p, nil
		}
	}
	return nil, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
