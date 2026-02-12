package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *Router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := rt.extractIdentifier(r)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	targetConvID := ps.ByName("targetConversationId")

	var req struct {
		MessageID string `json:"messageId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ✅ Use the new interface method
	content, image, err := rt.db.GetMessageByID(req.MessageID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	forwardedContent := "[Forwarded] " + content

	err = rt.db.SendMessage(targetConvID, userID, forwardedContent, image, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}