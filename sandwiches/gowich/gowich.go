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

	// "syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	jwt "github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

const defaultPort = "4007"

type shopOrder struct {
	Authorized bool   `json:"authorized"`
	Tenant     string `json:"tenant"`
	Runtime    string `json:"runtime"`
	Auth       string `json:"auth"`
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
	flagPort := flag.String("port", "", "Port for the Gowich to run on")

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
			log.Println("running command...")

			token, err := jwt.ParseWithClaims(r.Header["Token"][0], &shopOrder{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("there was an error")
				}
				return []byte(os.Getenv("JWT_SECRET")), nil
			})

			if err != nil {
				log.Println(err.Error())
				fmt.Fprintf(w, "There was an error")
				return
			}

			if claims, ok := token.Claims.(*shopOrder); ok && token.Valid {
				if chi.URLParam(r, "tenantID") == claims.Tenant {
					if !claims.Authorized {
						if len(claims.Auth) > 0 {
							out, err := placeOrder(claims.Auth, claims)
							if err != nil || out != "true" {
								log.Println(err.Error())
								fmt.Fprintf(w, "Not Authorized")
								return
							}
						} else {
							fmt.Fprintf(w, "Not Authorized")
							return
						}
					}

					order := chi.URLParam(r, "order")
					out, err := placeOrder(order, claims)

					if err != nil {
						log.Println(err.Error())
						fmt.Fprintf(w, "There was an error")
						return
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

func placeOrder(order string, claims *shopOrder) (string, error) {
	var cmd *exec.Cmd

	if strings.HasSuffix(order, ".go") {
		cmd = exec.Command(os.Getenv(strings.ToUpper(claims.Runtime)), "run", order)
	} else {
		cmd = exec.Command("./" + order)
	}

	tenants := "../tenants"
	if envTenants := os.Getenv("TENANTS"); envTenants != "" {
		tenants = envTenants
	}

	cmd.Dir = tenants + "/" + claims.Tenant

	// cmd.SysProcAttr = &syscall.SysProcAttr{}

	// uid, uerr := strconv.ParseUint(os.Getenv("UID"), 10, 32)
	// if uerr != nil {
	// 	log.Println(uerr.Error())
	// 	fmt.Fprintf(w, "There was an error")
	// 	return
	// }

	// gid, gerr := strconv.ParseUint(os.Getenv("GID"), 10, 32)
	// if gerr != nil {
	// 	log.Println(gerr.Error())
	// 	fmt.Fprintf(w, "There was an error")
	// 	return
	// }

	// hold := &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}

	out, err := cmd.Output()
	return string(out), err
}
