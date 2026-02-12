package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *Router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Authenticate sender
	senderID := rt.extractIdentifier(r)
	if senderID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 2. Get conversation ID from URL (now it's always the conversation ID)
	convID := ps.ByName("conversationId")
	if convID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 3. Parse multipart form (max 10 MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 4. Get text content (required)
	content := r.FormValue("content")
	if content == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "content is required"})
		return
	}

	// 5. Get optional image file
	var imageData []byte
	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		imageData, err = io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if err != http.ErrMissingFile {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 6. Get optional reply_to
	var replyTo *string
	if replyStr := r.FormValue("reply_to"); replyStr != "" {
		replyTo = &replyStr
	}

	// 7. Store message in database
	err = rt.db.SendMessage(convID, senderID, content, imageData, replyTo)
	if err != nil {
		// Check for specific error (e.g., user not in conversation)
		if err.Error() == "user not in conversation" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "you are not a member of this conversation"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}