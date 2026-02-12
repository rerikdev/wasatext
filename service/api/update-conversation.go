package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *Router) updateConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := rt.extractIdentifier(r)
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	convID := ps.ByName("conversationId")

	// ✅ Check if user is the creator
	creatorID, err := rt.db.GetConversationCreator(convID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if creatorID != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "only creator can update group"})
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	var photoData []byte
	file, _, err := r.FormFile("photo")
	if err == nil {
		defer file.Close()
		photoData, err = io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if err != http.ErrMissingFile {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = rt.db.UpdateConversation(convID, name, photoData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}