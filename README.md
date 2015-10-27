# Golang HTTP Handler

HTTP Handler for [Graphl Go](https://github.com/chris-ramon/graphql-go)

It has support only for POST requests of content type `application/json`.  
It has [GraphiQL](https://github.com/graphql/graphiql) inbuilt.

## Usage

```go
package main

import (
	"net/http"
	"github.com/sogko/graphql-go-handler"
)

func main() {
        gqlhandler.Init(schema.Load(), map[string]interface{}{
		"property_1": "user_1",
		"ctx":  "my_ctx_object",
	})
	http.HandleFunc("/graphql", gqlhandler.HandleGraphQL)
	http.HandleFunc("/graphiql", gqlhandler.HandleGraphiQL)
	port := ":3001"
	println("GraphQL server starting up on http://localhost" + port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic("ListenAndServe failed " + err.Error())
	}
}
```

### Details

```go
gqlhandler.Init(schema GraphQLSchema, rootObject map[string]interface{})

//The rootObject will be available in all your resolve functions in your Schemas like this
// You can get it like this
Resolve: func(p types.GQLFRParams) interface{} {
rootmap := p.Info.RootValue.(map[string]interface{})
}
```

The handler will accept requests with
the parameters:

  * **`query`**: A string GraphQL document to be executed.

  * **`variables`**: The runtime values to use for any GraphQL query variables
    as a JSON object.

  * **`operationName`**: If the provided `query` contains multiple named
    operations, this specifies which operation should be executed. If not
    provided, an 400 error will be returned if the `query` contains multiple
    named operations.

GraphQL will look in the POST request body.
The `handler` will interpret it
depending on the provided `Content-Type` header.

  * **`application/json`**: the POST body will be parsed as a JSON
    object of parameters.

  * **`application/graphql`**: The POST body will be parsed as GraphQL
    query string, which provides the `query` parameter.
