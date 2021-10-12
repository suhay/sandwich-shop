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

	"github.com/suhay/sandwich-shop/models"

	"github.com/go-chi/chi"
	jwt "github.com/golang-jwt/jwt"
)

// OrderContext hold values for the requested tenant and the order name
type OrderContext struct {
	TenantID string
	Order    models.Order
}

var OrderCtxKey = &contextKey{"order"}

// PlaceOrder takes the posted request and returns the shop value
func PlaceOrder(w http.ResponseWriter, r *http.Request) {
	// send order to waiting shop
	ctx := r.Context()
	res := Resolver{}
	order := OrderFromContext(ctx)
	user := UserFromContext(ctx)

	limit := 1
	s, serr := res.Query().Shops(ctx, *order.Order.Runtime, &limit)
	if serr != nil {
		panic(serr)
	}

	host := s[0].Host
	if s[0].Port != nil {
		host = host + ":" + strconv.Itoa(*s[0].Port)
	}

	urlParts := []string{host, order.TenantID, *order.Order.Path}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": user.Authorized,
		"tenant":     order.TenantID,
		"exp":        time.Now().Add(time.Minute * 1).Unix(),
		"runtime":    order.Order.Runtime,
		"auth":       order.Order.Auth,
	})

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

// SetOrderCtx set the OrderContext based upon incoming request values
func SetOrderCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := chi.URLParam(r, "tenantID")
		order := chi.URLParam(r, "order")

		ctx := r.Context()
		res := Resolver{}

		user := User{ID: tenantID, Authorized: false}
		ctxWithUser := context.WithValue(ctx, UserCtxKey, user)

		q, err := res.Query().Order(ctxWithUser, order)

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

		orderCtx := OrderContext{
			TenantID: tenantID,
			Order:    *q,
		}

		ctxWithOrder := context.WithValue(ctxWithUser, OrderCtxKey, orderCtx)
		next.ServeHTTP(w, r.WithContext(ctxWithOrder))
	})
}

// OrderFromContext returns the stored order context
func OrderFromContext(ctx context.Context) *OrderContext {
	raw, _ := ctx.Value(OrderCtxKey).(OrderContext)
	return &raw
}
