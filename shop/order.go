package sandwich_shop

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	jwt "github.com/golang-jwt/jwt"
)

// IncOrderContext hold values for the requested tenant and the order name
type IncOrderContext struct {
	TenantID string
	Order    string
}

type contextKey string

var incOrderKey = contextKey("incOrder")

// PlaceOrder takes the posted request and returns the shop value
func PlaceOrder(w http.ResponseWriter, r *http.Request) {
	// send order to waiting shop
	ctx := r.Context()
	res := Resolver{}
	incOrder := OrderContext(ctx)

	q, err := res.Query().Order(ctx, incOrder.Order)
	if err != nil {
		type Errors struct {
			Errors []string
		}
		errors := Errors{
			Errors: []string{err.Error()},
		}

		w.WriteHeader(http.StatusUnauthorized)
		b, err := json.Marshal(errors)
		if err != nil {
			panic(err)
		}
		w.Write(b)
		return
	}

	limit := 1
	s, serr := res.Query().Shops(ctx, *q.Runtime, &limit)
	if serr != nil {
		panic(serr)
	}

	host := s[0].Host
	if s[0].Port != nil {
		host = host + ":" + strconv.Itoa(*s[0].Port)
	}

	urlParts := []string{host, incOrder.TenantID, *q.Path}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
		"tenant":     incOrder.TenantID,
		"exp":        time.Now().Add(time.Minute * 1).Unix(),
		"runtime":    *q.Runtime,
	})
	log.Println(os.Getenv("JWT_SECRET"))
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		jwterr := fmt.Errorf("something went wrong validating the request: %s", err.Error())
		log.Println(jwterr)
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("POST", strings.Join(urlParts, "/"), r.Body)
	req.Header.Set("Token", tokenString)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		jwterr := fmt.Errorf("something went wrong sending client request: %s", err.Error())
		log.Println(jwterr)
		w.Write([]byte(`{ "error": "The shops seem to be down." }`))
		return
	}

	if body, _ := ioutil.ReadAll(resp.Body); len(body) > 0 {
		w.Header().Add("Content-Type", "application/json")
		w.Write(body)
		return
	}

	w.Write([]byte("{}"))
}

// SetOrderCtx set the IncOrderContext based upon incoming request values
func SetOrderCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := chi.URLParam(r, "tenantID")
		order := chi.URLParam(r, "order")

		incOrder := IncOrderContext{
			TenantID: tenantID,
			Order:    order,
		}

		ctx := context.WithValue(r.Context(), incOrderKey, incOrder)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OrderContext returns the stored order context
func OrderContext(ctx context.Context) *IncOrderContext {
	raw, _ := ctx.Value(incOrderKey).(IncOrderContext)
	return &raw
}
