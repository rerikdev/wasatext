package api

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	
)

func (rt *Router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	myID := rt.extractIdentifier(r)
	messageID := ps.ByName("messageId")

	err := rt.db.DeleteMessage(messageID, myID)
	if err != nil {
		w.WriteHeader(http.StatusForbidden) // 403: Not your message
		return
	}
	w.WriteHeader(http.StatusNoContent) // 204
}