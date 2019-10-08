package sandwich_shop

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/suhay/sandwich-shop/auth"
	yaml "gopkg.in/yaml.v2"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

// Resolver struct
type Resolver struct{}

// Query is the resolver bound to the type Query
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// OrderConfig represents the function configuration
type OrderConfig struct {
	Runtime string    `yaml:"runtime"`
	Path    string    `yaml:"path"`
	Env     []*string `yaml:"env"`
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) GetOrder(ctx context.Context, name string) (*Order, error) {
	user := auth.ForContext(ctx)
	if user == nil || (user != nil && user.ID == "") {
		return &Order{}, fmt.Errorf("Access denied")
	}

	orderConfig := make(map[string]OrderConfig)
	data, err := ioutil.ReadFile("../tenants/" + user.ID + "/orders.yml")
	if err != nil {
		return &Order{}, fmt.Errorf("error: %v", err)
	}

	err = yaml.Unmarshal([]byte(data), &orderConfig)
	if err != nil {
		return &Order{}, fmt.Errorf("error: %v", err)
	}

	path := orderConfig[name].Path
	env := orderConfig[name].Env
	runtime := Runtime(orderConfig[name].Runtime)

	return &Order{
		Name:    name,
		Runtime: &runtime,
		Path:    &path,
		Env:     env,
	}, nil
}

func (r *queryResolver) GetShops(ctx context.Context, name Runtime, limit *int) ([]*Shop, error) {
	user := auth.ForContext(ctx)
	if user == nil || (user != nil && user.ID == "") {
		return []*Shop{}, fmt.Errorf("Access denied")
	}

	avilableShops := []*Shop{}

	if mongodbURL := os.Getenv("MONGODB_URL"); len(mongodbURL) > 0 {
		panic("Not implemented!")
	} else {
		shops := []Shop{}
		dat, err := ioutil.ReadFile("shops.json")
		if err != nil {
			return []*Shop{}, fmt.Errorf("error: %v", err)
		}

		json.Unmarshal([]byte(dat), &shops)
		for i := range shops {
			for _, v := range shops[i].Runtimes {
				if *v == name {
					avilableShops = append(avilableShops, &shops[i])
					if *limit <= 1 || len(avilableShops) >= *limit {
						return avilableShops, nil
					}
				}
			}
		}
	}

	return avilableShops, nil
}
