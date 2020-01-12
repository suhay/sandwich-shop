package main

import (
  "log"
  "net/http"
  "os"
	"time"
	"flag"

  sandwich_shop "github.com/suhay/sandwich-shop"
  "github.com/suhay/sandwich-shop/auth"

  "github.com/99designs/gqlgen/handler"
  "github.com/go-chi/chi"
  "github.com/go-chi/chi/middleware"
  "github.com/joho/godotenv"
)

const defaultPort = "3002"

func setPort(port1, port2 string) string {
	if port1 != "" {
		return port1
	} else if port2 != "" {
		return port2
	}
	return defaultPort
}

func main() {
	flagEnvPath := flag.String("env", "", "Path to .env file to use")
	flagPort := flag.String("port", "", "Port for the Gowich to run one")

	flag.Parse()

	if *flagEnvPath != "" {
		if err := godotenv.Load(*flagEnvPath); err != nil {
			log.Println("Error loading .env file, defaulting to local files.")
		}
	} else {
		if err := godotenv.Load(); err != nil {
			log.Println("Error loading .env file, defaulting to local files.")
		}
	}

	port := setPort(*flagPort, os.Getenv("PORT"))

  r := chi.NewRouter()

  r.Use(middleware.RequestID)
  r.Use(middleware.Logger)
  r.Use(middleware.Recoverer)
  r.Use(middleware.Timeout(60 * time.Second))
  r.Use(middleware.ThrottleBacklog(2, 5, time.Second*61))

  r.Get("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Welcome to the Sandwich Shop!"))
  })

  r.Handle("/graphql", handler.Playground("GraphQL playground", "/query"))
  r.Handle("/query",
    handler.GraphQL(sandwich_shop.NewExecutableSchema(sandwich_shop.Config{Resolvers: &sandwich_shop.Resolver{}})),
  )

  r.Route("/shop", func(r chi.Router) {
    r.Route("/{tenantID}/{order}", func(r chi.Router) {
      r.Use(auth.Middleware)
      r.Use(sandwich_shop.SetOrderCtx)
      r.Post("/", sandwich_shop.PlaceOrder)
    })
  })

  log.Printf("connect to http://localhost:%s/ for GraphQL playground.", port)
  log.Fatal(http.ListenAndServe(":"+port, r))
}
