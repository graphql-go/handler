package handler

import (
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/graphql-go/graphql"

	"context"

	"github.com/graphql-go/graphql/gqlerrors"
)

const (
	ContentTypeJSON              = "application/json"
	ContentTypeGraphQL           = "application/graphql"
	ContentTypeFormURLEncoded    = "application/x-www-form-urlencoded"
	ContentTypeMultipartFormData = "multipart/form-data"
)

type MultipartFile struct {
	File   multipart.File
	Header *multipart.FileHeader
}

type ResultCallbackFn func(ctx context.Context, params *graphql.Params, result *graphql.Result, responseBody []byte)

type Handler struct {
	Schema           *graphql.Schema
	pretty           bool
	graphiql         bool
	playground       bool
	rootObjectFn     RootObjectFn
	resultCallbackFn ResultCallbackFn
	maxMemory        int64
	formatErrorFn    func(err error) gqlerrors.FormattedError
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

func getFromForm(values url.Values) *RequestOptions {
	query := values.Get("query")
	if query != "" {
		// get variables map
		variables := make(map[string]interface{}, len(values))
		variablesStr := values.Get("variables")
		json.Unmarshal([]byte(variablesStr), &variables)

		return &RequestOptions{
			Query:         query,
			Variables:     variables,
			OperationName: values.Get("operationName"),
		}
	}

	return nil
}

// RequestOptions Parses a http.Request into GraphQL request options struct
func NewRequestOptions(r *http.Request, maxMemory int64) *RequestOptions {
	if reqOpt := getFromForm(r.URL.Query()); reqOpt != nil {
		return reqOpt
	}

	if r.Method != http.MethodPost {
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
		if err := r.ParseForm(); err != nil {
			return &RequestOptions{}
		}

		if reqOpt := getFromForm(r.PostForm); reqOpt != nil {
			return reqOpt
		}

		return &RequestOptions{}

	case ContentTypeMultipartFormData:
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			// fmt.Printf("Parse Multipart Failed %v", err)
			return &RequestOptions{}
		}

		// @TODO handle array case...

		operationsParam := r.FormValue("operations")
		var opts RequestOptions
		if err := json.Unmarshal([]byte(operationsParam), &opts); err != nil {
			// fmt.Printf("Parse Operations Failed %v", err)
			return &RequestOptions{}
		}

		mapParam := r.FormValue("map")
		mapValues := make(map[string]([]string))
		if len(mapParam) != 0 {
			if err := json.Unmarshal([]byte(mapParam), &mapValues); err != nil {
				// fmt.Printf("Parse map Failed %v", err)
				return &RequestOptions{}
			}
		}

		variables := opts

		for key, value := range mapValues {
			for _, v := range value {
				if file, header, err := r.FormFile(key); err == nil {

					// Now set the path in ther variables
					var node interface{} = variables

					parts := strings.Split(v, ".")
					last := parts[len(parts)-1]

					for _, vv := range parts[:len(parts)-1] {
						// fmt.Printf("Doing vv=%s type=%T parts=%v\n", vv, node, parts)
						switch node.(type) {
						case RequestOptions:
							if vv == "variables" {
								node = opts.Variables
							} else {
								// panic("Invalid top level tag")
								return &RequestOptions{}
							}
						case map[string]interface{}:
							node = node.(map[string]interface{})[vv]
						case []interface{}:
							if idx, err := strconv.ParseInt(vv, 10, 64); err == nil {
								node = node.([]interface{})[idx]
							} else {
								// panic("Unable to lookup index")
								return &RequestOptions{}
							}
						default:
							// panic(fmt.Errorf("Unknown type %T", node))
							return &RequestOptions{}
						}
					}

					data := &MultipartFile{File: file, Header: header}

					switch node.(type) {
					case map[string]interface{}:
						node.(map[string]interface{})[last] = data
					case []interface{}:
						if idx, err := strconv.ParseInt(last, 10, 64); err == nil {
							node.([]interface{})[idx] = data
						} else {
							// panic("Unable to lookup index")
							return &RequestOptions{}
						}
					default:
						// panic(fmt.Errorf("Unknown last type %T", node))
						return &RequestOptions{}
					}
				}
			}
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

// ContextHandler provides an entrypoint into executing graphQL queries with a
// user-provided context.
func (h *Handler) ContextHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// get query
	opts := NewRequestOptions(r, h.maxMemory)

	// execute graphql query
	params := graphql.Params{
		Schema:         *h.Schema,
		RequestString:  opts.Query,
		VariableValues: opts.Variables,
		OperationName:  opts.OperationName,
		Context:        ctx,
	}
	if h.rootObjectFn != nil {
		params.RootObject = h.rootObjectFn(ctx, r)
	}
	result := graphql.Do(params)

	if formatErrorFn := h.formatErrorFn; formatErrorFn != nil && len(result.Errors) > 0 {
		formatted := make([]gqlerrors.FormattedError, len(result.Errors))
		for i, formattedError := range result.Errors {
			formatted[i] = formatErrorFn(formattedError.OriginalError())
		}
		result.Errors = formatted
	}

	if h.graphiql {
		acceptHeader := r.Header.Get("Accept")
		_, raw := r.URL.Query()["raw"]
		if !raw && !strings.Contains(acceptHeader, "application/json") && strings.Contains(acceptHeader, "text/html") {
			renderGraphiQL(w, params)
			return
		}
	}

	if h.playground {
		acceptHeader := r.Header.Get("Accept")
		_, raw := r.URL.Query()["raw"]
		if !raw && !strings.Contains(acceptHeader, "application/json") && strings.Contains(acceptHeader, "text/html") {
			renderPlayground(w, r)
			return
		}
	}

	// use proper JSON Header
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	var buff []byte
	if h.pretty {
		w.WriteHeader(http.StatusOK)
		buff, _ = json.MarshalIndent(result, "", "\t")

		w.Write(buff)
	} else {
		w.WriteHeader(http.StatusOK)
		buff, _ = json.Marshal(result)

		w.Write(buff)
	}

	if h.resultCallbackFn != nil {
		h.resultCallbackFn(ctx, &params, result, buff)
	}
}

// ServeHTTP provides an entrypoint into executing graphQL queries.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ContextHandler(r.Context(), w, r)
}

// RootObjectFn allows a user to generate a RootObject per request
type RootObjectFn func(ctx context.Context, r *http.Request) map[string]interface{}

type Config struct {
	Schema           *graphql.Schema
	Pretty           bool
	GraphiQL         bool
	Playground       bool
	RootObjectFn     RootObjectFn
	ResultCallbackFn ResultCallbackFn
	MaxMemory        int64
	FormatErrorFn    func(err error) gqlerrors.FormattedError
}

func NewConfig() *Config {
	return &Config{
		Schema:    nil,
		Pretty:    true,
		GraphiQL:  true,
		MaxMemory: 0,
	}
}

func New(p *Config) *Handler {
	if p == nil {
		p = NewConfig()
	}

	if p.Schema == nil {
		panic("undefined GraphQL schema")
	}

	maxMemory := p.MaxMemory
	if maxMemory == 0 {
		maxMemory = 32 << 20 // 32MB
	}

	return &Handler{
		Schema:           p.Schema,
		pretty:           p.Pretty,
		graphiql:         p.GraphiQL,
		playground:       p.Playground,
		rootObjectFn:     p.RootObjectFn,
		resultCallbackFn: p.ResultCallbackFn,
		maxMemory:        maxMemory,
		formatErrorFn:    p.FormatErrorFn,
	}
}
