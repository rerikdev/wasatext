package main

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// applyCORSHandler applies a CORS policy to the router. CORS stands for Cross-Origin Resource Sharing: it's a security
// feature present in web browsers that blocks JavaScript requests going across different domains if not specified in a
// policy. This function sends the policy of this API server.
func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{
			"x-example-header",
			"Content-Type",
			"Authorization",
			"Accept",
			"Origin",
			"X-Requested-With",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT", "PATCH"}),
		// Se vuoi permettere solo il frontend in sviluppo, usa:
		// handlers.AllowedOrigins([]string{"http://localhost:5173"}),
		// Altrimenti "*" permette tutte le origini (non sicuro in produzione)
		handlers.AllowedOrigins([]string{"*"}),
		handlers.MaxAge(1),
	)(h)
}
