package handler

import (
	"fmt"
	"html/template"
	"net/http"
)

type EditorCursorShape string

const (
	Line      EditorCursorShape = "line"
	Block     EditorCursorShape = "block"
	Underline EditorCursorShape = "underline"
)

type EditorTheme string

const (
	Dark  EditorTheme = "dark"
	Light EditorTheme = "light"
)

type RequestCredentials string

const (
	Omit       RequestCredentials = "omit"
	Include    RequestCredentials = "include"
	SameOrigin RequestCredentials = "same-origin"
)

type PlaygroundSettings struct {
	EditorCursorShape           EditorCursorShape  `json:"editor.cursorShape,omitempty"`
	EditorFontFamily            string             `json:"editor.fontFamily,omitempty"`
	EditorFontSize              float64            `json:"editor.fontSize,omitempty"`
	EditorReuseHeaders          bool               `json:"editor.reuseHeaders,omitempty"`
	EditorTheme                 EditorTheme        `json:"editor.theme,omitempty"`
	GeneralBetaUpdates          bool               `json:"general.betaUpdates,omitempty"`
	PrettierPrintWidth          float64            `json:"prettier.printWidth,omitempty"`
	PrettierTabWidth            float64            `json:"prettier.tabWidth,omitempty"`
	PrettierUseTabs             bool               `json:"prettier.useTabs,omitempty"`
	RequestCredentials          RequestCredentials `json:"request.credentials,omitempty"`
	RequestGlobalHeaders        map[string]string  `json:"request.globalHeaders,omitempty"`
	SchemaPollingEnable         bool               `json:"schema.polling.enable,omitempty"`
	SchemaPollingEndpointFilter string             `json:"schema.polling.endpointFilter,omitempty"`
	SchemaPollingInterval       float64            `json:"schema.polling.interval,omitempty"`
	SchemaDisableComments       bool               `json:"schema.disableComments,omitempty"`
	TracingHideTracingResponse  bool               `json:"tracing.hideTracingResponse,omitempty"`
	TracingTracingSupported     bool               `json:"tracing.tracingSupported,omitempty"`
}

type playgroundData struct {
	PlaygroundVersion    string
	Endpoint             string
	SubscriptionEndpoint string
	SetTitle             bool
	Settings             *PlaygroundSettings
}

// renderPlayground renders the Playground GUI
func (h *Handler) renderPlayground(w http.ResponseWriter, r *http.Request) {
	t := template.New("Playground")
	t, err := t.Parse(graphcoolPlaygroundTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var endpoint string
	if h.playground.Endpoint != "" {
		// in case the endpoint was explicitly set in the configuration use it here
		endpoint = h.playground.Endpoint
	} else {
		// in case no endpoint was specified assume the graphql api is served under the request's url
		endpoint = r.URL.Path
	}

	d := playgroundData{
		PlaygroundVersion:    graphcoolPlaygroundVersion,
		Endpoint:             endpoint,
		SubscriptionEndpoint: fmt.Sprintf("ws://%v/subscriptions", r.Host),
		SetTitle:             true,
		Settings:             h.playground.Settings,
	}

	err = t.ExecuteTemplate(w, "index", d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	return
}

const graphcoolPlaygroundVersion = "1.5.2"

const graphcoolPlaygroundTemplate = `
{{ define "index" }}
<!--
The request to this GraphQL server provided the header "Accept: text/html"
and as a result has been presented Playground - an in-browser IDE for
exploring GraphQL.

If you wish to receive JSON, provide the header "Accept: application/json" or
add "&raw" to the end of the URL within a browser.
-->
<!DOCTYPE html>
<html>

<head>
  <meta charset=utf-8/>
  <meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
  <title>GraphQL Playground</title>
  <link rel="stylesheet" href="//cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
  <link rel="shortcut icon" href="//cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
  <script src="//cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>

<body>
  <div id="root">
    <style>
      body {
        background-color: rgb(23, 42, 58);
        font-family: Open Sans, sans-serif;
        height: 90vh;
      }
      #root {
        height: 100%;
        width: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
      }
      .loading {
        font-size: 32px;
        font-weight: 200;
        color: rgba(255, 255, 255, .6);
        margin-left: 20px;
      }
      img {
        width: 78px;
        height: 78px;
      }
      .title {
        font-weight: 400;
      }
    </style>
    <img src='//cdn.jsdelivr.net/npm/graphql-playground-react/build/logo.png' alt=''>
    <div class="loading"> Loading
      <span class="title">GraphQL Playground</span>
    </div>
  </div>
  <script>window.addEventListener('load', function (event) {
      GraphQLPlayground.init(document.getElementById('root'), {
        // options as 'endpoint' belong here
        endpoint: {{ .Endpoint }},
        subscriptionEndpoint: {{ .SubscriptionEndpoint }},
        setTitle: {{ .SetTitle }},
		settings: {{ .Settings }}
      })
    })</script>
</body>

</html>
{{ end }}
`
