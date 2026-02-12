package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *Router) addParticipant(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := rt.extractIdentifier(r)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	convID := ps.ByName("conversationId")

	// ✅ Check if current user is an active participant (or creator)
	isParticipant, err := rt.db.IsParticipant(convID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isParticipant {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Convert username to user ID
	newUserID, err := rt.db.GetUserByName(req.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if newUserID == "" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}

	// Add the new participant
	err = rt.db.AddParticipant(convID, newUserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}