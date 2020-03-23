package session

import (
	"context"
	"net/http"
)

func AuthorizeRest(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r)
		return
		// w.WriteHeader(http.StatusUnauthorized)
	}
}

func AuthorizeGrpc(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
