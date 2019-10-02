package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"labix.org/v2/mgo/bson"
)

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

type requestBody struct {
	OperationName string `json:"operationName"`
}

// User is a stand-in for our database backed user object
type User struct {
	ID string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Middleware authorization for looking up user credentials
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if len(token) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		if body, _ := ioutil.ReadAll(r.Body); len(body) > 0 {
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			rBody := requestBody{}
			err := json.Unmarshal(body, &rBody)
			if err == nil {
				if rBody.OperationName == "IntrospectionQuery" {
					next.ServeHTTP(w, r)
					return
				}
			} else {
				log.Println("error: ", err)
			}
		}

		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file")
		}

		id := "b78682b3-36c8-4759-b8d1-5e62f029a1bc" // This will be replaced with checking the request

		if mongodbURL := os.Getenv("MONGODB_URL"); len(mongodbURL) > 0 {
			mctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			client, err := mongo.Connect(mctx, options.Client().SetRetryWrites(true).ApplyURI("mongodb+srv://"+os.Getenv("MONGODB_USER")+":"+os.Getenv("MONGODB_PASSWD")+"@"+mongodbURL+"?w=majority"))
			if err != nil {
				log.Fatal(err)
			}

			log.Println("Connecting to MongoDB")

			var result struct {
				Key string
			}

			filter := bson.M{"_id": id}
			client.Database("sandwich-shop").Collection("tenants").FindOne(mctx, filter).Decode(&result)

			// Check if the stored key equals the header
			if token == "Bearer "+result.Key {
				user := User{ID: id}
				ctx := context.WithValue(r.Context(), userCtxKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			log.Println("Invalid bearer token provided")
		} else {
			dat, err := ioutil.ReadFile("./tenants/" + id + "/.key")
			check(err)
			// Check if the stored key equals the header
			if token == "Bearer "+string(dat) {
				user := User{ID: id}
				ctx := context.WithValue(r.Context(), userCtxKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			log.Println("Invalid bearer token provided")
		}

		next.ServeHTTP(w, r)
	})
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func ForContext(ctx context.Context) *User {
	raw, _ := ctx.Value(userCtxKey).(User)
	return &raw
}
