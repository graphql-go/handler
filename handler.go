package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/schema"
	"github.com/graphql-go/graphql"
	"github.com/unrolled/render"
)

const (
	ContentTypeJSON           = "application/json"
	ContentTypeGraphQL        = "application/graphql"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
)

var decoder = schema.NewDecoder()

type Handler struct {
	Schema *graphql.Schema
	render *render.Render
}
type RequestOptions struct {
	Query         string                 `json:"query" url:"query" schema:"query"`
	Variables     map[string]interface{} `json:"variables" url:"variables" schema:"variables"`
	OperationName string                 `json:"operationName" url:"operationName" schema:"operationName"`
}

// a workaround for getting`variables` as a JSON string
type requestOptionsCompatibility struct {
	Query         string `json:"query" url:"query" schema:"query"`
	Variables     string `json:"variables" url:"variables" schema:"variables"`
	OperationName string `json:"operationName" url:"operationName" schema:"operationName"`
}

// RequestOptions Parses a http.Request into GraphQL request options struct
func NewRequestOptions(r *http.Request) *RequestOptions {

	query := r.URL.Query().Get("query")
	if query != "" {

		// get variables map
		var variables map[string]interface{}
		variablesStr := r.URL.Query().Get("variables")
		json.Unmarshal([]byte(variablesStr), variables)

		return &RequestOptions{
			Query:         query,
			Variables:     variables,
			OperationName: r.URL.Query().Get("operationName"),
		}
	}
	if r.Method != "POST" {
		return &RequestOptions{}
	}
	if r.Body == nil {
		return &RequestOptions{}
	}

	// TODO: improve Content-Type handling
	contentTypeStr := r.Header.Get("Content-Type")
	contentTypeTokens := strings.Split(contentTypeStr, ";")
	contentType := contentTypeTokens[0]

	switch contentType {
	case ContentTypeGraphQL:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return &RequestOptions{}
		}
		return &RequestOptions{
			Query: string(body),
		}
	case ContentTypeFormURLEncoded:
		var opts RequestOptions
		err := r.ParseForm()
		if err != nil {
			return &RequestOptions{}
		}
		err = decoder.Decode(&opts, r.PostForm)
		if err != nil {
			return &RequestOptions{}
		}
		return &opts
	case ContentTypeJSON:
		fallthrough
	default:
		var opts RequestOptions
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return &opts
		}
		err = json.Unmarshal(body, &opts)
		if err != nil {
			// Probably `variables` was sent as a string instead of an object.
			// So, we try to be polite and try to parse that as a JSON string
			var optsCompatible requestOptionsCompatibility
			json.Unmarshal(body, &optsCompatible)
			json.Unmarshal([]byte(optsCompatible.Variables), &opts.Variables)
		}
		return &opts
	}
}

// ServeHTTP provides an entry point into executing graphQL queries
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// get query
	opts := NewRequestOptions(r)

	// execute graphql query
	params := graphql.Params{
		Schema:         *h.Schema,
		RequestString:  opts.Query,
		VariableValues: opts.Variables,
		OperationName:  opts.OperationName,
	}
	result := graphql.Do(params)

	// render result
	h.render.JSON(w, http.StatusOK, result)
}

type Config struct {
	Schema *graphql.Schema
	Pretty bool
}

func NewConfig() *Config {
	return &Config{
		Schema: nil,
		Pretty: true,
	}
}

func New(p *Config) *Handler {
	if p == nil {
		p = NewConfig()
	}
	if p.Schema == nil {
		panic("undefined GraphQL schema")
	}
	r := render.New(render.Options{
		IndentJSON: p.Pretty,
	})
	return &Handler{
		Schema: p.Schema,
		render: r,
	}
}
