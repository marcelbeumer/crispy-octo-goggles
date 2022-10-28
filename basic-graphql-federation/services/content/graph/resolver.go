package graph

import (
	"fmt"

	"github.com/marcelbeumer/go-playground/basic-graphql-federation/services/content/graph/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	TopProducts []*model.Product
}

func NewResolver() *Resolver {
	products := make([]*model.Product, 10)
	for x := 0; x < len(products); x++ {
		products[x] = &model.Product{
			Sku: fmt.Sprintf("12345_%d", x),
		}
	}
	return &Resolver{TopProducts: products}
}
