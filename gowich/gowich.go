package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
	Runtime    string `json:"runtime"`
	jwt.StandardClaims
}

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
	i, _ := strconv.ParseInt(os.Getenv("TIMEOUT"), 10, 32)
	r.Use(middleware.Timeout(time.Second * time.Duration(i)))
	r.Use(middleware.ThrottleBacklog(2, 5, time.Second*time.Duration(i+1)))

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
					cmd := exec.Command(os.Getenv(strings.ToUpper(claims.Runtime)), "run", chi.URLParam(r, "order"))
					
					tenants := "../tenants"
					if envTenants := os.Getenv("TENANTS"); envTenants != "" {
						tenants = envTenants
					}

					cmd.Dir = tenants+"/"+claims.Tenant

					out, err := cmd.Output()
					if err != nil {
						log.Println(err.Error())
						fmt.Fprintf(w, "There was an error")
					}
					fmt.Fprintf(w, "%s", out)
					return
				}

				fmt.Fprintf(w, "Not Authorized")
				return
			}

			log.Println(err.Error())
			fmt.Fprintf(w, "There was an error")
			return
		}

		fmt.Fprintf(w, "Not Authorized")
	})

	log.Printf("Gowich online: http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
