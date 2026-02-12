package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *Router) getParticipants(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := rt.extractIdentifier(r)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	convID := ps.ByName("conversationId")

	// ✅ Check if user is a participant using the interface method
	isParticipant, err := rt.db.IsParticipant(convID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isParticipant {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	participants, err := rt.db.GetParticipants(convID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(participants)
}