package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ptflp/gotemporal/docs"
)

func registerDocsRoutes(r chi.Router) {
	r.Get("/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(docs.OpenAPI)
	})

	r.Get("/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(redocHTML))
	})
}

const redocHTML = `<!DOCTYPE html>
<html>
  <head>
    <title>Order API Docs</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="icon" href="data:,">
    <style>
      body { margin: 0; padding: 0; }
      redoc { height: 100vh; }
    </style>
  </head>
  <body>
    <redoc spec-url="/openapi.yaml"></redoc>
    <script src="https://cdn.jsdelivr.net/npm/redoc@2.3.0/bundles/redoc.standalone.js"></script>
  </body>
</html>`
