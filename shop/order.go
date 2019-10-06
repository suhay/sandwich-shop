package main

import (
	"encoding/json"
	"log"
	"net/http"

	sandwich_shop "github.com/suhay/sandwich-shop"
)

// PlaceOrder takes the posted request and returns the shop value
func PlaceOrder(w http.ResponseWriter, r *http.Request) {
	// check for order information
	// check for shop information??
	// send order to waiting shop
	ctx := r.Context()
	res := sandwich_shop.Resolver{}
	order, _ := ctx.Value(orderCtxKey).(string)

	q, err := res.Query().GetOrder(ctx, order)
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
	}

	log.Println(q)
}
