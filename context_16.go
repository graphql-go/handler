
// +build !go1.7

package handler

import (
	"net/http"
	"golang.org/x/net/context"
)

func contextFromRequest(req *http.Request) context.Context {
	// not yet supported, so fallback to background
	return context.Background()
}
