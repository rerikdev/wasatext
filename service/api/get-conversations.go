package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *Router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Authenticate user
	userID := rt.extractIdentifier(r)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 2. Fetch conversations from database
	conversations, err := rt.db.GetConversations(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. Return JSON
	w.Header().Set("Content-Type", "application/json")
	if conversations == nil {
		_ = json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	_ = json.NewEncoder(w).Encode(conversations)
}