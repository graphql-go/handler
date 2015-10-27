package gqlhandler

import (
	"bytes"
	"github.com/chris-ramon/graphql-go/testutil"
	"net/http"
	"reflect"
	"testing"
)

func TestRequestOptions_POST_ContentTypeApplicationGraphQL(t *testing.T) {
	body := []byte(`query RebelsShipsQuery { rebels { name } }`)
	expected := &requestOptions{
		Query: "query RebelsShipsQuery { rebels { name } }",
	}

	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBuffer(body))
	req.Header.Add("Content-Type", "application/graphql")
	result, err := getRequestOptions(req)

	if err != nil {
		t.Fatalf("Error Occurred %s", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestRequestOptions_POST_ContentTypeApplicationGraphQL_WithNonGraphQLQueryContent(t *testing.T) {
	body := []byte(`not a graphql query`)
	expected := &requestOptions{
		Query: "not a graphql query",
	}

	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBuffer(body))
	req.Header.Add("Content-Type", "application/graphql")
	result, err := getRequestOptions(req)

	if err != nil {
		t.Fatalf("Error Occurred %s", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestRequestOptions_POST_ContentTypeApplicationGraphQL_EmptyBody(t *testing.T) {
	body := []byte(``)
	expected := &requestOptions{
		Query: "",
	}
	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBuffer(body))
	req.Header.Add("Content-Type", "application/graphql")
	result, _ := getRequestOptions(req)
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestRequestOptions_POST_ContentTypeApplicationGraphQL_NilBody(t *testing.T) {
	req, _ := http.NewRequest("POST", "/graphql", nil)
	req.Header.Add("Content-Type", "application/graphql")
	_, err := getRequestOptions(req)

	if err.Error() != "The request body should not be empty" {
		t.Fatalf("A Wrong error occurred %s", err)
	}
}

func TestRequestOptions_POST_ContentTypeApplicationJSON(t *testing.T) {
	body := `
	{
		"query": "query RebelsShipsQuery { rebels { name } }"
	}`
	expected := &requestOptions{
		Query: "query RebelsShipsQuery { rebels { name } }",
	}

	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBufferString(body))
	req.Header.Add("Content-Type", "application/json")
	result, err := getRequestOptions(req)

	if err != nil {
		t.Fatalf("Error Occurred %s", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestRequestOptions_POST_ContentTypeApplicationJSON_WithVariablesAsObject(t *testing.T) {
	body := `
	{
		"query": "query RebelsShipsQuery { rebels { name } }",
		"variables": { "a": 1, "b": "2" }
	}`
	expected := &requestOptions{
		Query: "query RebelsShipsQuery { rebels { name } }",
		Variables: map[string]interface{}{
			"a": float64(1),
			"b": "2",
		},
	}

	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBufferString(body))
	req.Header.Add("Content-Type", "application/json")
	result, err := getRequestOptions(req)

	if err != nil {
		t.Fatalf("Error Occurred %s", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestRequestOptions_POST_ContentTypeApplicationJSON_WithVariablesAsString(t *testing.T) {
	body := `
	{
		"query": "query RebelsShipsQuery { rebels { name } }",
		"variables": "{ \"a\": 1, \"b\": \"2\" }"
	}`
	expected := &requestOptions{
		Query: "query RebelsShipsQuery { rebels { name } }",
		Variables: map[string]interface{}{
			"a": float64(1),
			"b": "2",
		},
	}

	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBufferString(body))
	req.Header.Add("Content-Type", "application/json")
	result, err := getRequestOptions(req)

	if err != nil {
		t.Fatalf("Error Occurred %s", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestRequestOptions_POST_ContentTypeApplicationJSON_WithInvalidJSON(t *testing.T) {
	body := `INVALIDJSON{}`
	expected := &requestOptions{}

	req, _ := http.NewRequest("POST", "/graphql", bytes.NewBufferString(body))
	req.Header.Add("Content-Type", "application/json")
	result, err := getRequestOptions(req)

	if err != nil {
		t.Fatalf("Error Occurred %s", err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestRequestOptions_POST_ContentTypeApplicationJSON_WithNilBody(t *testing.T) {
	req, _ := http.NewRequest("POST", "/graphql", nil)
	req.Header.Add("Content-Type", "application/json")
	_, err := getRequestOptions(req)
	if err.Error() != "The request body should not be empty" {
		t.Fatalf("A Wrong error occurred %s", err)
	}
}

func TestRequestOptions_POST_ContentTypeApplicationUrlEncoded_WithNilBody(t *testing.T) {
	req, _ := http.NewRequest("POST", "/graphql", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	_, err := getRequestOptions(req)
	if err.Error() != "The request body should not be empty" {
		t.Fatalf("A Wrong error occurred %s", err)
	}
}
