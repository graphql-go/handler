# graphql-go-handler [![CircleCI](https://circleci.com/gh/graphql-go/handler.svg?style=svg)](https://circleci.com/gh/graphql-go/handler) [![GoDoc](https://godoc.org/graphql-go/handler?status.svg)](https://godoc.org/github.com/graphql-go/handler) [![Coverage Status](https://coveralls.io/repos/graphql-go/handler/badge.svg?branch=master&service=github)](https://coveralls.io/github/graphql-go/handler?branch=master) [![Join the chat at https://gitter.im/graphql-go/graphql](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/graphql-go/graphql?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)


Golang HTTP.Handler for [graphl-go](https://github.com/graphql-go/graphql)

### Usage

```go
package main

import (
	"net/http"
	"github.com/graphql-go/handler"
)

func main() {
	schema, _ := graphql.NewSchema(...)

	h := handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
}
```

### Using Playground
```go
h := handler.New(&handler.Config{
	Schema: &schema,
	Pretty: true,
	GraphiQL: false,
	Playground: true,
})
```

### Creating a Custom Context

```go
package main

import (
	"context"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

type graphQLKey string

func makeMeField() *graphql.Field {
	userType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "User",
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.String,
				},
				"name": &graphql.Field{
					Type: graphql.String,
				},
			},
		},
	)
	return &graphql.Field{
		Type: userType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return p.Context.Value(graphQLKey("currentUser")), nil
		},
	}
}

func main() {
	queryType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"me": makeMeField(),
			},
		},
	)
	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
	graphQLHandler := handler.New(&handler.Config{
		Schema: &schema,
	})
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{1, "cool user"}
		ctx := context.WithValue(r.Context(), graphQLKey("currentUser"), user)
		graphQLHandler.ContextHandler(ctx, w, r)
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
}
```

### Details

The handler will accept requests with
the parameters:

  * **`query`**: A string GraphQL document to be executed.

  * **`variables`**: The runtime values to use for any GraphQL query variables
    as a JSON object.

  * **`operationName`**: If the provided `query` contains multiple named
    operations, this specifies which operation should be executed. If not
    provided, an 400 error will be returned if the `query` contains multiple
    named operations.

GraphQL will first look for each parameter in the URL's query-string:

```
/graphql?query=query+getUser($id:ID){user(id:$id){name}}&variables={"id":"4"}
```

If not found in the query-string, it will look in the POST request body.
The `handler` will interpret it
depending on the provided `Content-Type` header.

  * **`application/json`**: the POST body will be parsed as a JSON
    object of parameters.

  * **`application/x-www-form-urlencoded`**: this POST body will be
    parsed as a url-encoded string of key-value pairs.

  * **`application/graphql`**: The POST body will be parsed as GraphQL
    query string, which provides the `query` parameter.


### Examples
- [golang-graphql-playground](https://github.com/graphql-go/playground)
- [golang-relay-starter-kit](https://github.com/sogko/golang-relay-starter-kit)
- [todomvc-relay-go](https://github.com/sogko/todomvc-relay-go)

### Test
```bash
$ go get github.com/graphql-go/handler
$ go build && go test ./...
