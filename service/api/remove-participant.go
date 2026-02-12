package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *Router) removeParticipant(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := rt.extractIdentifier(r)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	convID := ps.ByName("conversationId")
	targetUserID := ps.ByName("userId")

	// User leaving themselves
	if targetUserID == userID {
		err := rt.db.RemoveParticipant(convID, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Check if current user is the creator
	creatorID, err := rt.db.GetConversationCreator(convID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if creatorID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = rt.db.RemoveParticipant(convID, targetUserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}