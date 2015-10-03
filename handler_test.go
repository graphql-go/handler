package gqlhandler_test

import (
	"encoding/json"
	"fmt"
	"github.com/chris-ramon/graphql-go/testutil"
	"github.com/chris-ramon/graphql-go/types"
	"github.com/sogko/graphql-go-handler"
	"github.com/sogko/graphql-relay-go/examples/starwars"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func decodeResponse(t *testing.T, recorder *httptest.ResponseRecorder) *types.GraphQLResult {
	// clone request body reader so that we can have a nicer error message
	bodyString := ""
	var target types.GraphQLResult
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
func executeTest(t *testing.T, h *gqlhandler.Handler, req *http.Request) (*types.GraphQLResult, *httptest.ResponseRecorder) {
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	result := decodeResponse(t, resp)
	return result, resp
}

func TestHandler_BasicQuery(t *testing.T) {

	expected := &types.GraphQLResult{
		Data: map[string]interface{}{
			"rebels": map[string]interface{}{
				"id":   "RmFjdGlvbjox",
				"name": "Alliance to Restore the Republic",
			},
		},
	}
	queryString := `query=query RebelsShipsQuery { rebels { id, name } }`
	req, _ := http.NewRequest("GET", fmt.Sprintf("/graphql?%v", queryString), nil)

	h := gqlhandler.New(&gqlhandler.Config{
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
func TestHandler_Params_NilParams(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if str, ok := r.(string); ok {
				if str != "undefined graphQL schema" {
					t.Fatalf("unexpected error, got %v", r)
				}
				// test passed
				return
			}
			t.Fatalf("unexpected error, got %v", r)

		}
		t.Fatalf("expected to panic, did not panic")
	}()
	_ = gqlhandler.New(nil)

}
