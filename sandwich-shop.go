package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/suhay/sandwich-shop/auth"
	"github.com/suhay/sandwich-shop/models"
	shop "github.com/suhay/sandwich-shop/shop"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
)

const defaultPort = "3002"

func setPort(port1 string) string {
	if port1 != "" {
		return port1
	} else if val, ok := os.LookupEnv("PORT"); ok {
		return val
	}
	return defaultPort
}

func setMode(mode1 string) *string {
	if mode1 != "" {
		return &mode1
	} else if val, ok := os.LookupEnv("MODE"); ok {
		return &val
	}
	return nil
}

func main() {
	flagEnvPath := flag.String("env", "", "Path to .env file to use")
	flagPort := flag.String("port", "", "Port for the Sandwich Shop to open up on")
	flagMode := flag.String("mode", "", "Mode flag (DEV)")

	flag.Parse()

	if *flagEnvPath != "" {
		if err := godotenv.Load(*flagEnvPath); err != nil {
			log.Println("Error loading .env file, defaulting to local files....")
		}
	} else {
		if err := godotenv.Load(); err != nil {
			log.Println("Error loading .env file, defaulting to local files..")
		}
	}

	port := setPort(*flagPort)
	srv := handler.NewDefaultServer(models.NewExecutableSchema(models.Config{Resolvers: &shop.Resolver{}}))
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.ThrottleBacklog(2, 5, time.Second*61))

	if mode := setMode(*flagMode); mode != nil {
		if *mode == "DEV" {
			log.Println("Running in development mode.")
			r.Handle("/graphql", playground.Handler("GraphQL playground", "/query"))
		}
	}

	r.Handle("/query", srv)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the Sandwich Shop!"))
	})

	r.Route("/shop", func(r chi.Router) {
		r.Route("/{tenantID}/{order}", func(r chi.Router) {
			r.Use(shop.SetOrderCtx)
			r.Use(auth.Middleware)
			r.Post("/", shop.PlaceOrder)
		})
	})

	log.Printf("Sandwich Shop open at http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
