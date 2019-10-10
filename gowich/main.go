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

type shopOrder struct {
	Authorized bool   `json:"authorized"`
	Tenant     string `json:"tenant"`
	jwt.StandardClaims
}

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

	r.Post("/{tenantID}/{order}", func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jwt.ParseWithClaims(r.Header["Token"][0], &shopOrder{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return []byte(os.Getenv("JWT_SECRET")), nil
			})
			if err != nil {
				log.Println(err.Error())
				fmt.Fprintf(w, "There was an error")
			}

			if claims, ok := token.Claims.(*shopOrder); ok && token.Valid {
				if chi.URLParam(r, "tenantID") == claims.Tenant && claims.Authorized {
					out, err := exec.Command("go", "run", "../tenants/"+claims.Tenant+"/"+chi.URLParam(r, "order")).Output()
					if err != nil {
						log.Println(err.Error())
						fmt.Fprintf(w, "There was an error")
					}
					fmt.Fprintf(w, "%s", out)
					return
				}
			} else {
				log.Println(err.Error())
				fmt.Fprintf(w, "There was an error")
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
