package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/marcelbeumer/go-playground/basic-graphql-federation/services/content/graph/generated"
	"github.com/marcelbeumer/go-playground/basic-graphql-federation/services/content/graph/model"
)

// FindProductBySku is the resolver for the findProductBySku field.
func (r *entityResolver) FindProductBySku(ctx context.Context, sku string) (*model.Product, error) {
	panic(fmt.Errorf("not implemented: FindProductBySku - findProductBySku"))
}

// Entity returns generated.EntityResolver implementation.
func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
