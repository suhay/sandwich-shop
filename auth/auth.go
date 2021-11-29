package auth

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	shop "github.com/suhay/sandwich-shop/shop"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"labix.org/v2/mgo/bson"
)

// Middleware authorization for looking up user credentials
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		order := shop.OrderFromContext(ctx)

		if order.Order.Auth != nil && len(*order.Order.Auth) > 0 {
			user := shop.User{ID: order.TenantID, Authorized: false}
			ctxWithUser := context.WithValue(ctx, shop.UserCtxKey, user)
			next.ServeHTTP(w, r.WithContext(ctxWithUser))
			return
		}

		var token string
		if authHeader := order.Order.AuthHeader; authHeader != nil && *authHeader != "" {
			token = r.Header.Get(*authHeader)
		} else {
			token = strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", -1)
		}

		if len(token) == 0 {
			log.Println("Authorization token not found.")
			w.WriteHeader(http.StatusUnauthorized)
			next.ServeHTTP(w, r)
			return
		}

		var storedKey string
		var authError error

		if mongodbURL, ok := os.LookupEnv("MONGODB_URL"); ok {
			storedKey, authError = mongoAuth(order.TenantID, mongodbURL)
		} else {
			storedKey, authError = localAuth(order.TenantID)
		}

		if authError != nil {
			w.WriteHeader(http.StatusInternalServerError)
			next.ServeHTTP(w, r)
			return
		}

		// Check if the stored key equals the header
		if token == storedKey {
			user := shop.User{ID: order.TenantID, Authorized: true}
			ctxWithUser := context.WithValue(ctx, shop.UserCtxKey, user)
			next.ServeHTTP(w, r.WithContext(ctxWithUser))
			return
		}

		log.Println("Invalid bearer token provided")
		w.WriteHeader(http.StatusForbidden)
		next.ServeHTTP(w, r)
	})
}

func mongoAuth(tenantID string, mongodbURL string) (string, error) {
	mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if cancel != nil {
		defer cancel()
	}

	log.Println("Connecting to MongoDB")
	client, err := mongo.Connect(mctx, options.Client().SetRetryWrites(true).ApplyURI("mongodb+srv://"+os.Getenv("MONGODB_USER")+":"+os.Getenv("MONGODB_PASSWD")+"@"+mongodbURL+"/?w=majority"))
	if err != nil {
		log.Println(err)
		return "", err
	}

	log.Println("Connected")

	var result struct {
		Key string
	}

	filter := bson.M{"_id": tenantID}
	client.Database("sandwich-shop").Collection("tenants").FindOne(mctx, filter).Decode(&result)
	defer client.Disconnect(mctx)

	return result.Key, nil
}

func localAuth(tenantID string) (string, error) {
	tenants := "tenants"
	if envTenants, ok := os.LookupEnv("TENANTS"); ok {
		tenants = envTenants
	}

	dat, err := os.ReadFile(tenants + "/" + tenantID + "/.key")
	if err != nil {
		log.Println(err)
		return "", err
	}

	return strings.TrimSpace(string(dat)), nil
}
