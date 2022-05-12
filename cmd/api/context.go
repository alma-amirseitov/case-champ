package main

import (
	"case-championship/internal/data"
	"context"
	"net/http"
)

type contextKey string

const userContextKey = contextKey("admin")

func (app *application) contextSetAdmin(r *http.Request, user *data.Admin) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetAdmin(r *http.Request) *data.Admin {
	user, ok := r.Context().Value(userContextKey).(*data.Admin)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
