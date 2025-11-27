// File: cmd/api/context.go
package main

import (
	"context"
	"net/http"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

// contextSetUser adds the user information to the request context.
func (app *app) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user) // Add user to context
	return r.WithContext(ctx)                                   // Return a new request with the updated context
}

// contextGetUser retrieves the user information from the request context.
func (app *app) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User) // Retrieve user from context
	if !ok {
		panic("missing user value in context") // Panic if user is not found in context
	}
	return user // Return the retrieved user
}
