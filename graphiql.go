package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/graphql-go/graphql"
)

// data is the page data structure of the rendered GraphiQL page
type graphiqlData struct {
	GraphiqlVersion              string
	SubscriptionTransportVersion string
	QueryString                  string
	ResultString                 string
	VariablesString              string
	OperationName                string
	Endpoint                     template.URL
	SubscriptionEndpoint         template.URL
	UsingHTTP                    bool
	UsingWS                      bool
}

// renderGraphiQL renders the GraphiQL GUI
func renderGraphiQL(w http.ResponseWriter, params graphql.Params, handler Handler) {
	t := template.New("GraphiQL")
	t, err := t.Parse(graphiqlTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create variables string
	vars, err := json.MarshalIndent(params.VariableValues, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	varsString := string(vars)
	if varsString == "null" {
		varsString = ""
	}

	// Create result string
	var resString string
	if params.RequestString == "" {
		resString = ""
	} else {
		result, err := json.MarshalIndent(graphql.Do(params), "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resString = string(result)
	}

	isEndpointUsingWS := strings.HasPrefix(handler.Endpoint, "ws://")
	UsingHTTP := !isEndpointUsingWS
	UsingWS := isEndpointUsingWS || handler.SubscriptionsEndpoint != ""
	SubscriptionEndpoint := ""
	if UsingWS {
		if isEndpointUsingWS {
			SubscriptionEndpoint = handler.Endpoint
		} else {
			SubscriptionEndpoint = handler.SubscriptionsEndpoint
		}
	}

	d := graphiqlData{
		GraphiqlVersion:              graphiqlVersion,
		SubscriptionTransportVersion: subscriptionTransportVersion,
		QueryString:                  params.RequestString,
		ResultString:                 resString,
		VariablesString:              varsString,
		OperationName:                params.OperationName,
		Endpoint:                     template.URL(handler.Endpoint),
		SubscriptionEndpoint:         template.URL(SubscriptionEndpoint),
		UsingHTTP:                    UsingHTTP,
		UsingWS:                      UsingWS,
	}
	err = t.ExecuteTemplate(w, "index", d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	return
}

// graphiqlVersion is the current version of GraphiQL
const graphiqlVersion = "0.11.11"

// subscriptionTransportVersion is the current version of the subscription transport of GraphiQL
const subscriptionTransportVersion = "0.8.2"

// tmpl is the page template to render GraphiQL
const graphiqlTemplate = `
{{ define "index" }}
<!--
The request to this GraphQL server provided the header "Accept: text/html"
and as a result has been presented GraphiQL - an in-browser IDE for
exploring GraphQL.

If you wish to receive JSON, provide the header "Accept: application/json" or
add "&raw" to the end of the URL within a browser.
-->
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <title>GraphiQL</title>
  <meta name="robots" content="noindex" />
  <meta name="referrer" content="origin">
  <style>
    body {
      height: 100%;
      margin: 0;
      overflow: hidden;
      width: 100%;
    }
    #graphiql {
      height: 100vh;
    }
  </style>
  <link href="//cdn.jsdelivr.net/npm/graphiql@{{ .GraphiqlVersion }}/graphiql.css" rel="stylesheet" />
  <script src="//cdn.jsdelivr.net/es6-promise/4.0.5/es6-promise.auto.min.js"></script>
  <script src="//cdn.jsdelivr.net/fetch/0.9.0/fetch.min.js"></script>
  <script src="//cdn.jsdelivr.net/react/15.4.2/react.min.js"></script>
  <script src="//cdn.jsdelivr.net/react/15.4.2/react-dom.min.js"></script>
  <script src="//cdn.jsdelivr.net/npm/graphiql@{{ .GraphiqlVersion }}/graphiql.min.js"></script>

  {{ if .UsingHTTP }}
    <script src="//cdn.jsdelivr.net/fetch/2.0.1/fetch.min.js"></script>
  {{ end }}
  {{ if .UsingWS }}
    <script src="//unpkg.com/subscriptions-transport-ws@{{ .SubscriptionTransportVersion }}/browser/client.js"></script>
  {{ end }}
  {{ if and .UsingWS .UsingHTTP }}
    <script src="//unpkg.com/graphiql-subscriptions-fetcher@0.0.2/browser/client.js"></script>
  {{ end }}

</head>
<body>
  <div id="graphiql">Loading...</div>
  <script>
    // Collect the URL parameters
    var parameters = {};
    window.location.search.substr(1).split('&').forEach(function (entry) {
      var eq = entry.indexOf('=');
      if (eq >= 0) {
        parameters[decodeURIComponent(entry.slice(0, eq))] =
          decodeURIComponent(entry.slice(eq + 1));
      }
    });

    // Produce a Location query string from a parameter object.
    function locationQuery(params, location) {
      return (location ? location: '') + '?' + Object.keys(params).map(function (key) {
        return encodeURIComponent(key) + '=' +
          encodeURIComponent(params[key]);
      }).join('&');
    }

    // Derive a fetch URL from the current URL, sans the GraphQL parameters.
    var graphqlParamNames = {
      query: true,
      variables: true,
      operationName: true
    };

    var otherParams = {};
    for (var k in parameters) {
      if (parameters.hasOwnProperty(k) && graphqlParamNames[k] !== true) {
        otherParams[k] = parameters[k];
      }
    }

    {{ if .UsingWS }}
      var subscriptionsClient = new window.SubscriptionsTransportWs.SubscriptionClient({{ .SubscriptionEndpoint }}, {
        reconnect: true
      });
      var graphQLWSFetcher = subscriptionsClient.request.bind(subscriptionsClient);
    {{ end }}

    {{ if .UsingHTTP }}
      var fetchURL = locationQuery(otherParams, {{ .Endpoint }});

      // Defines a GraphQL fetcher using the fetch API.
      function graphQLHttpFetcher(graphQLParams) {
        return fetch(fetchURL, {
          method: 'post',
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(graphQLParams),
          credentials: 'include',
        }).then(function (response) {
          return response.text();
        }).then(function (responseBody) {
          try {
            return JSON.parse(responseBody);
          } catch (error) {
            return responseBody;
          }
        });
      }
    {{ end }}

    {{ if and .UsingWS .UsingHTTP }}
      var fetcher = window.GraphiQLSubscriptionsFetcher.graphQLFetcher(subscriptionsClient, graphQLHttpFetcher);
    {{ else }}
      {{ if .UsingWS }}
        var fetcher = graphQLWSFetcher;
      {{ end }}
      {{ if .UsingHTTP }}
        var fetcher = graphQLHttpFetcher;
      {{ end }}
    {{ end }}

    // When the query and variables string is edited, update the URL bar so
    // that it can be easily shared.
    function onEditQuery(newQuery) {
      parameters.query = newQuery;
      updateURL();
    }

    function onEditVariables(newVariables) {
      parameters.variables = newVariables;
      updateURL();
    }

    function onEditOperationName(newOperationName) {
      parameters.operationName = newOperationName;
      updateURL();
    }

    function updateURL() {
      history.replaceState(null, null, locationQuery(parameters));
    }

    // Render <GraphiQL /> into the body.
    ReactDOM.render(
      React.createElement(GraphiQL, {
        fetcher: fetcher,
        onEditQuery: onEditQuery,
        onEditVariables: onEditVariables,
        onEditOperationName: onEditOperationName,
        query: {{ .QueryString }},
        response: {{ .ResultString }},
        variables: {{ .VariablesString }},
        operationName: {{ .OperationName }},
      }),
      document.getElementById('graphiql')
    );
  </script>
</body>
</html>
{{ end }}
`
