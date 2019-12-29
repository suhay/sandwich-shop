package sandwich_shop

//go:generate go run github.com/99designs/gqlgen

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/suhay/sandwich-shop/auth"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		client, err := mongo.Connect(mctx, options.Client().SetRetryWrites(true).ApplyURI("mongodb+srv://"+os.Getenv("MONGODB_USER")+":"+os.Getenv("MONGODB_PASSWD")+"@"+mongodbURL+"/"))
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Connecting to MongoDB")

		var limit64 int64
		if limit == nil {
			defaultLimit := 10
			limit64 = int64(defaultLimit)
		} else {
			limit64 = int64(*limit)
		}

		findOptions := options.Find()
		findOptions.SetLimit(limit64)

		filter := bson.M{"runtimes": name}
		shops, err := client.Database("sandwich-shop").Collection("shops").Find(mctx, filter, findOptions)

		if cancel != nil {
			cancel()
		}

		if err != nil {
			log.Println(err)
			return []*Shop{}, fmt.Errorf("error: %v", err)
		}
		if shops == nil {
			return []*Shop{}, nil
		}

		for shops.Next(mctx) {
			var result *Shop
			shops.Decode(&result)
			avilableShops = append(avilableShops, result)
		}

		shops.Close(mctx)
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
