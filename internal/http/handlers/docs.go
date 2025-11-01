package handlers

import (
	_ "embed"
	"net/http"
)

// Embed the OpenAPI specification file from the handlers directory
//go:embed openapi.yaml
var openapiSpec []byte

// DocsHandler handles documentation-related HTTP endpoints
type DocsHandler struct{}

// NewDocsHandler creates a new documentation handler
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// ServeOpenAPISpec handles GET /openapi.yaml
// Returns the OpenAPI specification as YAML
func (h *DocsHandler) ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers to allow documentation to be accessed from different origins
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set content type for YAML
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	// Write the embedded OpenAPI spec
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(openapiSpec); err != nil {
		// Log error but don't change response as headers are already sent
		return
	}
}

// ServeRedocUI handles GET /docs
// Returns the Redoc documentation UI
func (h *DocsHandler) ServeRedocUI(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set content type for HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	// Generate Redoc HTML page
	html := generateRedocHTML()

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(html)); err != nil {
		// Log error but don't change response as headers are already sent
		return
	}
}

// generateRedocHTML returns the HTML template for Redoc UI
func generateRedocHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Gatekeeper API Documentation</title>
    <meta name="description" content="Gatekeeper Wallet-Native Authentication API - Interactive API documentation powered by Redoc">

    <!-- Redoc doesn't require any CSS files, it's all embedded -->
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
        }
    </style>
</head>
<body>
    <!-- Redoc container -->
    <div id="redoc-container"></div>

    <!-- Redoc standalone bundle -->
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>

    <!-- Initialize Redoc -->
    <script>
        // Initialize Redoc with our OpenAPI spec
        Redoc.init(
            '/openapi.yaml',
            {
                // Redoc options
                scrollYOffset: 0,
                hideDownloadButton: false,
                disableSearch: false,
                theme: {
                    colors: {
                        primary: {
                            main: '#3b82f6'
                        }
                    },
                    typography: {
                        fontSize: '15px',
                        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
                        headings: {
                            fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif'
                        }
                    },
                    sidebar: {
                        backgroundColor: '#f8fafc',
                        textColor: '#1e293b'
                    },
                    rightPanel: {
                        backgroundColor: '#1e293b'
                    }
                }
            },
            document.getElementById('redoc-container')
        );
    </script>
</body>
</html>`
}
