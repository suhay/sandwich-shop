package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const defaultPort = "4007"

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
		out, err := exec.Command("date").Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("The date is %s\n", out)
		// w.Write([]byte("Welcome to the Sandwich Shop!"))
	})

	log.Printf("connect to http://localhost:%s/ for GraphQL playground.", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
