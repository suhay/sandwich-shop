package sandwich_shop

//go:generate go run github.com/99designs/gqlgen

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	shop "github.com/suhay/sandwich-shop/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	yaml "gopkg.in/yaml.v2"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type contextKey struct {
	name string
}

// Resolver struct
type Resolver struct {
}

// Query is the resolver bound to the type Query
func (r *Resolver) Query() shop.QueryResolver {
	return &queryResolver{r}
}

// OrderConfig represents the function configuration
type OrderConfig struct {
	Runtime    string    `yaml:"runtime"`
	Path       string    `yaml:"path"`
	Env        []*string `yaml:"env"`
	Auth       string    `yaml:"auth"`
	AuthHeader string    `yaml:"auth_header"`
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Order(ctx context.Context, name string) (*shop.Order, error) {
	user := UserFromContext(ctx)
	if user == nil || (user != nil && user.ID == "") {
		return &shop.Order{}, fmt.Errorf("access denied")
	}

	tenants := "tenants"
	if envTenants, ok := os.LookupEnv("TENANTS"); ok {
		tenants = envTenants
	}

	orderConfig := make(map[string]OrderConfig)
	data, err := os.ReadFile(tenants + "/" + user.ID + "/orders.yml")
	if err != nil {
		return &shop.Order{}, fmt.Errorf("error: %v", err)
	}

	err = yaml.Unmarshal([]byte(data), &orderConfig)
	if err != nil {
		return &shop.Order{}, fmt.Errorf("error: %v", err)
	}

	path := orderConfig[name].Path
	env := orderConfig[name].Env
	runtime := shop.Runtime(orderConfig[name].Runtime)
	auth := orderConfig[name].Auth
	authHeader := orderConfig[name].AuthHeader

	return &shop.Order{
		Name:       name,
		Runtime:    &runtime,
		Path:       &path,
		Env:        env,
		Auth:       &auth,
		AuthHeader: &authHeader,
	}, nil
}

func (r *queryResolver) Shops(ctx context.Context, name shop.Runtime, limit *int) ([]*shop.Sandwich, error) {
	sandwiches, err := r.Sandwiches(ctx, name, limit)
	return sandwiches, err
}

func (r *queryResolver) Sandwiches(ctx context.Context, name shop.Runtime, limit *int) ([]*shop.Sandwich, error) {
	// Check for admin user when we build out the admin dashboards
	// user := auth.ForContext(ctx)
	// if user == nil || (user != nil && user.ID == "") {
	// 	return []*Shop{}, fmt.Errorf("access denied")
	// }

	availableSandwiches := []*shop.Sandwich{}
	var limit64 int64

	if limit == nil {
		defaultLimit := 10
		limit64 = int64(defaultLimit)
		limit = &defaultLimit
	} else {
		limit64 = int64(*limit)
	}

	if mongodbURL, ok := os.LookupEnv("MONGODB_URL"); ok {
		mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		client, err := mongo.Connect(mctx, options.Client().SetRetryWrites(true).ApplyURI("mongodb+srv://"+os.Getenv("MONGODB_USER")+":"+os.Getenv("MONGODB_PASSWD")+"@"+mongodbURL+"/"))
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Connecting to MongoDB")

		findOptions := options.Find()
		findOptions.SetLimit(limit64)

		filter := bson.M{"runtimes": name}
		sandwiches, err := client.Database("sandwich-shop").Collection("sandwiches").Find(mctx, filter, findOptions)

		if cancel != nil {
			defer cancel()
		}

		if err != nil {
			log.Println(err)
			return []*shop.Sandwich{}, fmt.Errorf("error: %v", err)
		}
		if sandwiches == nil {
			return []*shop.Sandwich{}, nil
		}

		for sandwiches.Next(mctx) {
			var result *shop.Sandwich
			sandwiches.Decode(&result)
			availableSandwiches = append(availableSandwiches, result)
		}

		defer sandwiches.Close(mctx)
	} else {
		tenants := "tenants"
		if envTenants, ok := os.LookupEnv("TENANTS"); ok {
			tenants = envTenants
		}

		sandwichesFilePath := tenants + "/sandwiches.json"
		if envSandwiches, ok := os.LookupEnv("SANDWICHES"); ok {
			sandwichesFilePath = envSandwiches
		}

		sandwiches := []shop.Sandwich{}
		dat, err := os.ReadFile(sandwichesFilePath)
		if err != nil {
			return []*shop.Sandwich{}, fmt.Errorf("error: %v", err)
		}

		json.Unmarshal([]byte(dat), &sandwiches)
		for i := range sandwiches {
			for _, v := range sandwiches[i].Runtimes {
				if *v == name {
					availableSandwiches = append(availableSandwiches, &sandwiches[i])
					if *limit <= 1 || len(availableSandwiches) >= *limit {
						return availableSandwiches, nil
					}
				}
			}
		}
	}

	return availableSandwiches, nil
}
