package sandwich_shop

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
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

	q, err := res.Query().GetOrder(ctx, incOrder.Order)
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

	limit := 1
	s, serr := res.Query().GetShops(ctx, *q.Runtime, &limit)
	if serr != nil {
		panic(serr)
	}

	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	log.Println(string(b))
	w.Write(b)
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
