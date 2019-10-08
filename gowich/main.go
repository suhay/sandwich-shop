package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
)

const defaultPort = "4007"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file, defaulting to local files.")
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.ThrottleBacklog(2, 5, time.Second*61))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return os.Getenv("JWT_SECRET"), nil
			})
			if err != nil {
				fmt.Fprintf(w, err.Error())
			}

			if token.Valid {
				out, err := exec.Command("date").Output()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Fprintf(w, "The date is %s\n", out)
				return
			}
		} else {
			fmt.Fprintf(w, "Not Authorized")
			return
		}

		w.Write([]byte("What would you like on your Gowich?"))
	})

	log.Printf("connect to http://localhost:%s/ for GraphQL playground.", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
