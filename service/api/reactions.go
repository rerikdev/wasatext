package api

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
)

func (rt *Router) addReaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	myID := rt.extractIdentifier(r)
	if myID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	messageID := ps.ByName("messageId")
	emoticon := ps.ByName("reactionId")

	err := rt.db.AddReaction(messageID, myID, emoticon)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent) // 204
}

func (rt *Router) removeReaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	myID := rt.extractIdentifier(r)
	if myID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	messageID := ps.ByName("messageId")
	emoticon := ps.ByName("reactionId")

	err := rt.db.RemoveReaction(messageID, myID, emoticon)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent) // 204
}