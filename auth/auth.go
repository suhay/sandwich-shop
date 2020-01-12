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

  "github.com/go-chi/chi"
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

// Middleware authorization for looking up user credentials
func Middleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    token := r.Header.Get("Authorization")
    if len(token) == 0 {
			log.Println("Authorization token not found.")
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

    tenantID := chi.URLParam(r, "tenantID")

    if mongodbURL := os.Getenv("MONGODB_URL"); len(mongodbURL) > 0 {
      mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
      if cancel != nil {
        defer cancel()
      }
      
			log.Println("Connecting to MongoDB")
      client, err := mongo.Connect(mctx, options.Client().SetRetryWrites(true).ApplyURI("mongodb+srv://"+os.Getenv("MONGODB_USER")+":"+os.Getenv("MONGODB_PASSWD")+"@"+mongodbURL+"/?w=majority"))
      if err != nil {
        log.Println(err)
      }
			log.Println("Connected")

      var result struct {
        Key string
      }

      filter := bson.M{"_id": tenantID}
      client.Database("sandwich-shop").Collection("tenants").FindOne(mctx, filter).Decode(&result)
			defer client.Disconnect(mctx)
			
      // Check if the stored key equals the header
      if token == "Bearer "+result.Key {
				user := User{ID: tenantID}
        ctx := context.WithValue(r.Context(), userCtxKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
        return
      }
      log.Println("Invalid bearer token provided")
    } else {
      dat, err := ioutil.ReadFile("../tenants/" + tenantID + "/.key")
      if err != nil {
        log.Println(err)
      }
      // Check if the stored key equals the header
      if token == "Bearer "+string(dat) {
        user := User{ID: tenantID}
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
