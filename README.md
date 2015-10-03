# graphql-go-handler

Golang HTTP.Handler for graphl-go

### Usage

```go
package main

import (
	"net/http"
	"github.com/sogko/graphql-go-handler"

)

func main() {

  // define GraphQL schema using relay library helpers
  schema := types.NewGraphQLSchema(...)
  
	h := gqlhandler.New(&gqlhandler.Config{
  		Schema: &starwars.Schema,
  		Pretty: true,
  })
	
	// serve HTTP
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


### Test
```bash
$ go get github.com/sogko/graphql-go-handler
$ go build && go test ./...
```