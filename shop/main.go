package main

import (
	"log"
	"net/http"
	"os"

	sandwich_shop "github.com/suhay/sandwich-shop"
	"github.com/suhay/sandwich-shop/auth"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
)

const defaultPort = "3002"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	router := chi.NewRouter()
	router.Use(auth.Middleware)
	// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("welcome"))
	// })

	router.Handle("/", handler.Playground("GraphQL playground", "/query"))
	router.Handle("/query",
		handler.GraphQL(sandwich_shop.NewExecutableSchema(sandwich_shop.Config{Resolvers: &sandwich_shop.Resolver{}})),
	)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground. PID: %d", port, os.Getpid())
	log.Fatal(http.ListenAndServe(":"+port, router))
}
