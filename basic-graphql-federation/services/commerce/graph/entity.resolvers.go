package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/marcelbeumer/go-playground/basic-graphql-federation/services/commerce/graph/generated"
	"github.com/marcelbeumer/go-playground/basic-graphql-federation/services/commerce/graph/model"
)

// FindProductBySku is the resolver for the findProductBySku field.
func (r *entityResolver) FindProductBySku(ctx context.Context, sku string) (*model.Product, error) {
	for _, p := range r.TopProducts {
		if p.Sku == sku {
			return p, nil
		}
	}
	return nil, nil
}

// Entity returns generated.EntityResolver implementation.
func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
