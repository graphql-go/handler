// +build go1.7

package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/testutil"
)

func TestContextPropagatedFromRequest(t *testing.T) {
	myNameQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Name: "name",
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Context.Value("name"), nil
				},
			},
		},
	})

	myNameSchema, err := graphql.NewSchema(graphql.SchemaConfig{Query: myNameQuery})
	if err != nil {
		t.Fatal(err)
	}

	expected := &graphql.Result{
		Data: map[string]interface{}{
			"name": "context-data",
		},
	}
	queryString := `query={name}`
	req, _ := http.NewRequest("GET", fmt.Sprintf("/graphql?%v", queryString), nil)

	h := New(&Config{Schema: &myNameSchema, Pretty: true})

	ctx := context.WithValue(req.Context(), "name", "context-data")
	*req = *req.WithContext(ctx)

	resp := httptest.NewRecorder()

	h.ServeHTTP(resp, req)

	result := decodeResponse(t, resp)

	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
