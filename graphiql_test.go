package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/graphql-go/graphql/testutil"
	"github.com/graphql-go/handler"
)

func TestRenderGraphiQL(t *testing.T) {
	cases := map[string]struct {
		graphiqlEnabled      bool
		accept               string
		url                  string
		expectedStatusCode   int
		expectedContentType  string
		expectedBodyContains string
	}{
		"renders GraphiQL": {
			graphiqlEnabled:      true,
			accept:               "text/html",
			expectedStatusCode:   http.StatusOK,
			expectedContentType:  "text/html; charset=utf-8",
			expectedBodyContains: "<!DOCTYPE html>",
		},
		"doesn't render graphiQL if turned off": {
			graphiqlEnabled:     false,
			accept:              "text/html",
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "application/json; charset=utf-8",
		},
		"doesn't render GraphiQL if Content-Type application/json is present": {
			graphiqlEnabled:     true,
			accept:              "application/json,text/html",
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "application/json; charset=utf-8",
		},
		"doesn't render GraphiQL if Content-Type text/html is not present": {
			graphiqlEnabled:     true,
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "application/json; charset=utf-8",
		},
		"doesn't render GraphiQL if 'raw' query is present": {
			graphiqlEnabled:     true,
			accept:              "text/html",
			url:                 "?raw",
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "application/json; charset=utf-8",
		},
	}

	for tcID, tc := range cases {
		t.Run(tcID, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tc.url, nil)
			if err != nil {
				t.Error(err)
			}

			req.Header.Set("Accept", tc.accept)

			h := handler.New(&handler.Config{
				Schema:   &testutil.StarWarsSchema,
				GraphiQL: tc.graphiqlEnabled,
			})

			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)
			resp := rr.Result()

			statusCode := resp.StatusCode
			if statusCode != tc.expectedStatusCode {
				t.Fatalf("%s: wrong status code, expected %v, got %v", tcID, tc.expectedStatusCode, statusCode)
			}

			contentType := resp.Header.Get("Content-Type")
			if contentType != tc.expectedContentType {
				t.Fatalf("%s: wrong content type, expected %s, got %s", tcID, tc.expectedContentType, contentType)
			}

			body := rr.Body.String()
			if !strings.Contains(body, tc.expectedBodyContains) {
				t.Fatalf("%s: wrong body, expected %s to contain %s", tcID, body, tc.expectedBodyContains)
			}
		})
	}
}
