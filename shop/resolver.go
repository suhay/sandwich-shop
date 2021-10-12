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
	// shops map[models.Runtime][]*models.Shop
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
	if envTenants := os.Getenv("TENANTS"); envTenants != "" {
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

func (r *queryResolver) Shops(ctx context.Context, name shop.Runtime, limit *int) ([]*shop.Shop, error) {
	// Check for admin user when we build out the admin dashboards
	// user := auth.ForContext(ctx)
	// if user == nil || (user != nil && user.ID == "") {
	// 	return []*Shop{}, fmt.Errorf("access denied")
	// }

	avilableShops := []*shop.Shop{}
	var limit64 int64

	if limit == nil {
		defaultLimit := 10
		limit64 = int64(defaultLimit)
		limit = &defaultLimit
	} else {
		limit64 = int64(*limit)
	}

	if mongodbURL := os.Getenv("MONGODB_URL"); len(mongodbURL) > 0 {
		mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		client, err := mongo.Connect(mctx, options.Client().SetRetryWrites(true).ApplyURI("mongodb+srv://"+os.Getenv("MONGODB_USER")+":"+os.Getenv("MONGODB_PASSWD")+"@"+mongodbURL+"/"))
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Connecting to MongoDB")

		findOptions := options.Find()
		findOptions.SetLimit(limit64)

		filter := bson.M{"runtimes": name}
		shops, err := client.Database("sandwich-shop").Collection("shops").Find(mctx, filter, findOptions)

		if cancel != nil {
			defer cancel()
		}

		if err != nil {
			log.Println(err)
			return []*shop.Shop{}, fmt.Errorf("error: %v", err)
		}
		if shops == nil {
			return []*shop.Shop{}, nil
		}

		for shops.Next(mctx) {
			var result *shop.Shop
			shops.Decode(&result)
			avilableShops = append(avilableShops, result)
		}

		defer shops.Close(mctx)
	} else {
		shopsFilePath := os.Getenv("SHOPS")

		shops := []shop.Shop{}
		dat, err := os.ReadFile(shopsFilePath)
		if err != nil {
			return []*shop.Shop{}, fmt.Errorf("error: %v", err)
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
