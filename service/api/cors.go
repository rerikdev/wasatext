package api

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
)

// CORS middleware wraps an httprouter.Handle and adds CORS headers.
func CORS(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Set CORS headers for every response
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "1") // 1 second as per spec

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler
		next(w, r, ps)
	}
}

// CORSWrapper can be used to wrap http.HandlerFunc if needed, but we use the above.