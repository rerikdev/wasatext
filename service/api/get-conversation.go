package api

import (
	"encoding/json"
	"net/http"
	"github.com/julienschmidt/httprouter"
)

func (rt *Router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := rt.extractIdentifier(r)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	param := ps.ByName("conversationId")
	var convID string
	var err error

	// Try to interpret as conversation ID
	convType, err := rt.db.GetConversationType(param)
	if err == nil && convType != "" {
		convID = param
	} else {
		// Resolve username to ID and get/create direct conversation
		theirID, err := rt.extractUser(param)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if theirID == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		convID, err = rt.db.GetDirectConversationID(userID, theirID)
		if err != nil {
			convID, err = rt.db.CreateConversation(userID, "direct", "", nil, []string{theirID})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	// ✅ Set the numeric conversation ID in a response header
	w.Header().Set("X-Conversation-Id", convID)

	// Fetch messages
	messages, err := rt.db.GetConversation(convID, userID)
	if err != nil {
		if err.Error() == "access denied or conversation not found" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = rt.db.MarkMessagesAsRead(convID, userID, nil)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(messages)
}