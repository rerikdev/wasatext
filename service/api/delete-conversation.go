package api

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
)

func (rt *Router) deleteConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Identify the requester
	myID := rt.extractIdentifier(r)
	if myID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 2. TRANSLATION FIX: Convert "Maria" -> "2"
	// If we don't do this, the DB tries to delete messages where user is "Maria" (which don't exist)
	theirNameOrID := ps.ByName("conversationId")
	theirID, err := rt.extractUser(theirNameOrID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. Execute the deletion
	err = rt.db.DeleteConversation(myID, theirID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}