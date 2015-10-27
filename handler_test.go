package gqlhandler_test

import (
	"bytes"
	"fmt"
	"github.com/sogko/graphql-go-handler"
	"github.com/sogko/graphql-relay-go/examples/starwars"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func executeTest(t *testing.T, req *http.Request, expected string) {
	resp := httptest.NewRecorder()
	gqlhandler.Init(starwars.Schema, map[string]interface{}{})
	gqlhandler.HandleGraphQL(resp, req)
	bodyString := ""
	if b, err := ioutil.ReadAll(resp.Body); err == nil {
		bodyString = string(b)
	}
	if bodyString != expected {
		t.Fatalf("\nExpected: %s \nActually: %s", expected, bodyString)
	}
}

func TestHandler_BasicQuery(t *testing.T) {
	queryString := `query=query RebelsShipsQuery { rebels { id, name } }`
	expected := "Only POST requests are allowed\n"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/graphql?%v", queryString), nil)
	executeTest(t, req, expected)
}

func TestHandler_BasicPost(t *testing.T) {
	body := `
	{
		"query": "query RebelsShipsQuery { rebels { name } }"
	}`
	expected := `{"data":{"rebels":{"name":"Alliance to Restore the Republic"}}}`
	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBufferString(body))
	req.Header.Add("Content-Type", "application/json")
	executeTest(t, req, expected)
}
