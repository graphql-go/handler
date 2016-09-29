
// +build go1.7

package handler

import (
	"context"
	"net/http"
)

func contextFromRequest(req *http.Request) context.Context {
	return req.Context()
}
