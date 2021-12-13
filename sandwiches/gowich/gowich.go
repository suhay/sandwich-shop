package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	// "syscall"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	jwt "github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

const defaultPort = "4007"

type order struct {
	Auth       string `json:"auth"`
	AuthHeader string `json:"auth_header"`
	Authorized bool   `json:"authorized"`
	Env        string `json:"env"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	Runtime    string `json:"runtime"`
	Tenant     string `json:"tenant"`
	jwt.StandardClaims
}

func setPort(port1 string) string {
	if port1 != "" {
		return port1
	} else if val, ok := os.LookupEnv("PORT"); ok {
		return val
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

	port := setPort(*flagPort)

	timeout := 60
	if val, ok := os.LookupEnv("TIMEOUT"); ok {
		if i, err := fmt.Sscan(val); err == nil {
			timeout = i
		}
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Duration(timeout) * time.Second))

	r.Post("/{tenantID}/{order}", func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jwt.ParseWithClaims(r.Header.Get("Token"), &order{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("there was an error")
				}
				return []byte(os.Getenv("JWT_SECRET")), nil
			})

			if err != nil {
				log.Println(err.Error())
				http.Error(w, http.StatusText(500), 500)
				return
			}

			if claims, ok := token.Claims.(*order); ok && token.Valid {
				if chi.URLParam(r, "tenantID") == claims.Tenant {
					rBody, _ := ioutil.ReadAll(r.Body)
					defer r.Body.Close()
					body := string(rBody)

					var header string
					if authHeader := claims.AuthHeader; authHeader != "" {
						header = r.Header.Get(authHeader)
					} else {
						header = r.Header.Get("Authorization")
					}

					if !claims.Authorized {
						if len(claims.Auth) > 0 {
							out, err := placeOrder(claims.Auth, claims, body, header)
							if err != nil || strings.ToLower(out) != "true" {
								http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
								return
							}
						} else {
							http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
							return
						}
					}

					order := chi.URLParam(r, "order")
					out, err := placeOrder(order, claims, body, header)

					if err != nil {
						log.Println(err.Error())
						http.Error(w, http.StatusText(500), 500)
						return
					}
					fmt.Fprintf(w, "%s", out)
					return
				}

				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			log.Println(err.Error())
			http.Error(w, http.StatusText(500), 500)
			return
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})

	log.Printf("Gowich online: http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func placeOrder(order string, claims *order, body string, header string) (string, error) {
	var cmd *exec.Cmd
	if claims.Runtime == "binary" {
		cmd = exec.Command(order, body, header)
	} else {
		runtime := strings.ToUpper(claims.Runtime)
		name := os.Getenv(runtime)

		if strings.HasPrefix(runtime, "GO") {
			cmd = exec.Command(name, "run", order, body, header)
		} else {
			cmd = exec.Command(name, order, body, header)
		}
	}

	if claims.Env != "[]" && len(claims.Env) > 0 {
		env := []string{}
		json.Unmarshal([]byte(claims.Env), &env)
		cmd.Env = env
	}

	tenants := "../../tenants"
	if envTenants, ok := os.LookupEnv("TENANTS"); ok {
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
	return strings.TrimSpace(string(out)), err
}
