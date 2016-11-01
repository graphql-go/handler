// +build !go1.7

package handler

import (
	"net/http"

	"golang.org/x/net/context"
)

// ServeHTTP provides an entrypoint into executing graphQL queries.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ContextHandler(context.Background(), w, r)
}
