package handler_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/testutil"
	"github.com/graphql-go/handler"
	"github.com/graphql-go/relay/examples/starwars" // TODO: remove this dependency
	"golang.org/x/net/context"
)

func decodeResponse(t *testing.T, recorder *httptest.ResponseRecorder) *graphql.Result {
	// clone request body reader so that we can have a nicer error message
	bodyString := ""
	var target graphql.Result
	if b, err := ioutil.ReadAll(recorder.Body); err == nil {
		bodyString = string(b)
	}
	readerClone := strings.NewReader(bodyString)

	decoder := json.NewDecoder(readerClone)
	err := decoder.Decode(&target)
	if err != nil {
		t.Fatalf("DecodeResponseToType(): %v \n%v", err.Error(), bodyString)
	}
	return &target
}
func executeTest(t *testing.T, h *handler.Handler, req *http.Request) (*graphql.Result, *httptest.ResponseRecorder) {
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	result := decodeResponse(t, resp)
	return result, resp
}

func TestContextPropagated(t *testing.T) {
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
	myNameSchema, err := graphql.NewSchema(graphql.SchemaConfig{myNameQuery, nil})
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

	h := handler.New(&handler.Config{
		Schema: &myNameSchema,
		Pretty: true,
	})

	ctx := context.WithValue(context.Background(), "name", "context-data")
	resp := httptest.NewRecorder()
	h.ContextHandler(ctx, resp, req)
	result := decodeResponse(t, resp)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestHandler_BasicQuery_Pretty(t *testing.T) {
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"rebels": map[string]interface{}{
				"id":   "RmFjdGlvbjox",
				"name": "Alliance to Restore the Republic",
			},
		},
	}
	queryString := `query=query RebelsShipsQuery { rebels { id, name } }`
	req, _ := http.NewRequest("GET", fmt.Sprintf("/graphql?%v", queryString), nil)

	h := handler.New(&handler.Config{
		Schema: &starwars.Schema,
		Pretty: true,
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestHandler_BasicQuery_Ugly(t *testing.T) {
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"rebels": map[string]interface{}{
				"id":   "RmFjdGlvbjox",
				"name": "Alliance to Restore the Republic",
			},
		},
	}
	queryString := `query=query RebelsShipsQuery { rebels { id, name } }`
	req, _ := http.NewRequest("GET", fmt.Sprintf("/graphql?%v", queryString), nil)

	h := handler.New(&handler.Config{
		Schema: &starwars.Schema,
		Pretty: false,
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestHandler_Params_NilParams(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if str, ok := r.(string); ok {
				if str != "undefined GraphQL schema" {
					t.Fatalf("unexpected error, got %v", r)
				}
				// test passed
				return
			}
			t.Fatalf("unexpected error, got %v", r)

		}
		t.Fatalf("expected to panic, did not panic")
	}()
	_ = handler.New(nil)

}
