package sandwich_shop

import (
	"context"
	"fmt"
	"log"

	"github.com/suhay/sandwich-shop/auth"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

// Resolver struct
type Resolver struct{}

// Query : Get the resolver bound to the type Query
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) GetOrder(ctx context.Context, name string) (*Order, error) {

	log.Println("Checking auth")

	if user := auth.ForContext(ctx); user == nil {
		return &Order{}, fmt.Errorf("Access denied")
	}

	log.Println("Auth done!")

	var path string
	var runtime Runtime

	path = "get_sandwich.js"
	runtime = RuntimeNode12_7

	log.Println("sending data back")

	return &Order{
		Name:    name,
		Runtime: &runtime,
		Path:    &path,
	}, nil
}
