package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/location"
	"github.com/graphql-go/graphql/testutil"
	"github.com/graphql-go/handler"
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

func uploadTest(t *testing.T, mapData string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	queryString := `{
        "query":"query HeroNameQuery { hero { name } }",
        "variables":{"file":[null,null]}
    }`

	writer.WriteField("operations", queryString)
	if mapData != "" {
		writer.WriteField("map", mapData)

		part1, _ := writer.CreateFormFile("0", "test1.txt")
		if _, err := io.Copy(part1, strings.NewReader("How now brown cow")); err != nil {
			t.Fatalf("unexpected copy writer fail %v", err)
		}
		part2, _ := writer.CreateFormFile("1", "test2.txt")
		if _, err := io.Copy(part2, strings.NewReader("How now gold fish")); err != nil {
			t.Fatalf("unexpected copy writer fail %v", err)
		}
	}

	err := writer.Close()
	if err != nil {
		t.Fatalf("unexpected writer fail %v", err)
	}

	req, err := http.NewRequest("POST", "/graphql", body)
	if err != nil {
		t.Fatalf("unexpected NewRequest fail %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
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
	myNameSchema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: myNameQuery,
	})
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
			"hero": map[string]interface{}{
				"name": "R2-D2",
			},
		},
	}
	queryString := `query=query HeroNameQuery { hero { name } }&operationName=HeroNameQuery`
	req, _ := http.NewRequest("GET", fmt.Sprintf("/graphql?%v", queryString), nil)

	callbackCalled := false
	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
		Pretty: true,
		ResultCallbackFn: func(ctx context.Context, params *graphql.Params, result *graphql.Result, responseBody []byte) {
			callbackCalled = true
			if params.OperationName != "HeroNameQuery" {
				t.Fatalf("OperationName passed to callback was not HeroNameQuery: %v", params.OperationName)
			}

			if result.HasErrors() {
				t.Fatalf("unexpected graphql result errors")
			}
		},
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
	if !callbackCalled {
		t.Fatalf("ResultCallbackFn was not called when it should have been")
	}
}

func TestHandler_BasicQuery_Ugly(t *testing.T) {
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"hero": map[string]interface{}{
				"name": "R2-D2",
			},
		},
	}
	queryString := `query=query HeroNameQuery { hero { name } }`
	req, _ := http.NewRequest("GET", fmt.Sprintf("/graphql?%v", queryString), nil)

	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
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

func TestHandler_BasicQuery_WithRootObjFn(t *testing.T) {
	myNameQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Name: "name",
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					rv := p.Info.RootValue.(map[string]interface{})
					return rv["rootValue"], nil
				},
			},
		},
	})
	myNameSchema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: myNameQuery,
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := &graphql.Result{
		Data: map[string]interface{}{
			"name": "foo",
		},
	}
	queryString := `query={name}`
	req, _ := http.NewRequest("GET", fmt.Sprintf("/graphql?%v", queryString), nil)

	h := handler.New(&handler.Config{
		Schema: &myNameSchema,
		Pretty: true,
		RootObjectFn: func(ctx context.Context, r *http.Request) map[string]interface{} {
			return map[string]interface{}{"rootValue": "foo"}
		},
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestHandler_Post(t *testing.T) {
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"hero": map[string]interface{}{
				"name": "R2-D2",
			},
		},
	}
	queryString := `{"query":"query HeroNameQuery { hero { name } }"}`

	req, _ := http.NewRequest("POST", "/graphql", strings.NewReader(queryString))
	req.Header.Set("Content-Type", "application/json")

	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestHandler_Multipart_Basic(t *testing.T) {
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"hero": map[string]interface{}{
				"name": "R2-D2",
			},
		},
	}

	req := uploadTest(t, "")

	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestHandler_Multipart_Basic_ErrNoOperation(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err := writer.Close()
	if err != nil {
		t.Fatalf("unexpected writer fail %v", err)
	}

	req, err := http.NewRequest("POST", "/graphql", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if len(result.Errors) != 1 || result.Errors[0].Message != "Must provide an operation." {
		t.Fatalf("unexpected response")
	}
}

func TestHandler_Multipart_Basic_ErrBadMap(t *testing.T) {
	req := uploadTest(t, `{`)

	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if len(result.Errors) != 1 || result.Errors[0].Message != "Must provide an operation." {
		t.Fatalf("unexpected response")
	}
}

func TestHandler_Multipart_Basic_ErrBadMapRoot(t *testing.T) {
	req := uploadTest(t, `{"0":["xxx.file"]}`)

	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if len(result.Errors) != 1 || result.Errors[0].Message != "Must provide an operation." {
		t.Fatalf("unexpected response %+v", result)
	}
}

func TestHandler_Multipart_Basic_Upload(t *testing.T) {
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"hero": map[string]interface{}{
				"name": "R2-D2",
			},
		},
	}

	req := uploadTest(t, `{"0":["variables.file"]}`)

	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestHandler_Multipart_Basic_UploadSlice(t *testing.T) {
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"hero": map[string]interface{}{
				"name": "R2-D2",
			},
		},
	}

	req := uploadTest(t, `{"0":["variables.file.0"],"1":["variables.file.1"]}`)

	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
type customError struct {
	message string
}

func (e customError) Error() string {
	return fmt.Sprintf("%s", e.message)
}

func TestHandler_BasicQuery_WithFormatErrorFn(t *testing.T) {
	resolverError := customError{message: "resolver error"}
	myNameQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Name: "name",
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, resolverError
				},
			},
		},
	})
	myNameSchema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: myNameQuery,
	})
	if err != nil {
		t.Fatal(err)
	}

	customFormattedError := gqlerrors.FormattedError{
		Message: resolverError.Error(),
		Locations: []location.SourceLocation{
			location.SourceLocation{
				Line:   1,
				Column: 2,
			},
		},
		Path: []interface{}{"name"},
	}

	expected := &graphql.Result{
		Data: map[string]interface{}{
			"name": nil,
		},
		Errors: []gqlerrors.FormattedError{customFormattedError},
	}

	queryString := `query={name}`
	req, _ := http.NewRequest("GET", fmt.Sprintf("/graphql?%v", queryString), nil)

	formatErrorFnCalled := false
	h := handler.New(&handler.Config{
		Schema: &myNameSchema,
		Pretty: true,
		FormatErrorFn: func(err error) gqlerrors.FormattedError {
			formatErrorFnCalled = true
			var formatted gqlerrors.FormattedError
			switch err := err.(type) {
			case *gqlerrors.Error:
				formatted = gqlerrors.FormatError(err)
			default:
				t.Fatalf("unexpected error type: %v", reflect.TypeOf(err))
			}
			return formatted
		},
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if !formatErrorFnCalled {
		t.Fatalf("FormatErrorFn was not called when it should have been")
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestHandler_Multipart_Basic_BadSlice(t *testing.T) {
	req := uploadTest(t, `{"0":["variables.file.x"]}`)

	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if len(result.Errors) != 1 || result.Errors[0].Message != "Must provide an operation." {
		t.Fatalf("unexpected response %+v", result)
	}
}

func TestHandler_Multipart_Basic_BadSliceLast(t *testing.T) {
	req := uploadTest(t, `{"0":["variables.file.0.test"]}`)

	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if len(result.Errors) != 1 || result.Errors[0].Message != "Must provide an operation." {
		t.Fatalf("unexpected response %+v", result)
	}
}

func TestHandler_Multipart_Basic_BadSliceMiddle(t *testing.T) {
	req := uploadTest(t, `{"0":["variables.file.x.test"]}`)

	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if len(result.Errors) != 1 || result.Errors[0].Message != "Must provide an operation." {
		t.Fatalf("unexpected response %+v", result)
	}
}

func TestHandler_Multipart_Basic_BadMapPath(t *testing.T) {
	req := uploadTest(t, `{"0":["variables.x.y.z.z.y"]}`)

	h := handler.New(&handler.Config{
		Schema: &testutil.StarWarsSchema,
	})
	result, resp := executeTest(t, h, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected server response %v", resp.Code)
	}
	if len(result.Errors) != 1 || result.Errors[0].Message != "Must provide an operation." {
		t.Fatalf("unexpected response %+v", result)
	}
}
