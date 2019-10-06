package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	sandwich_shop "github.com/suhay/sandwich-shop"
	"github.com/suhay/sandwich-shop/auth"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const defaultPort = "3002"

var tenantIDCtxKey = &contextKey{"tenantID"}
var orderCtxKey = &contextKey{"order"}

type contextKey struct {
	name string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

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
			r.Use(orderCtx)
			r.Use(auth.Middleware)
			r.Post("/", PlaceOrder)
		})
	})

	log.Printf("connect to http://localhost:%s/ for GraphQL playground. PID: %d", port, os.Getpid())
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func orderCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := chi.URLParam(r, "tenantID")
		order := chi.URLParam(r, "order")

		ctx := context.WithValue(r.Context(), tenantIDCtxKey, tenantID)
		ctx = context.WithValue(ctx, orderCtxKey, order)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
