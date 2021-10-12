package sandwich_shop

import "context"

// User is a stand-in for our database backed user object
type User struct {
	ID         string
	Authorized bool
}

var UserCtxKey = &contextKey{"user"}

// UserFromContext finds the user from the context. REQUIRES Middleware to have run.
func UserFromContext(ctx context.Context) *User {
	raw, _ := ctx.Value(UserCtxKey).(User)
	return &raw
}
