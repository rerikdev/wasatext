package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *Router) createConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	creatorID := rt.extractIdentifier(r)
	if creatorID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	convType := r.FormValue("type")
	if convType != "direct" && convType != "group" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "type must be 'direct' or 'group'"})
		return
	}

	var name string
	if convType == "group" {
		name = r.FormValue("name")
		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "group name is required"})
			return
		}
	}

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

	// Members are sent as JSON array of usernames
	membersJSON := r.FormValue("members")
	var memberUsernames []string
	if membersJSON != "" {
		if err := json.Unmarshal([]byte(membersJSON), &memberUsernames); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid members JSON"})
			return
		}
	}

	// Convert usernames to IDs
	var memberIDs []string
	for _, username := range memberUsernames {
		id, err := rt.db.GetUserByName(username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "user not found: " + username})
			return
		}
		memberIDs = append(memberIDs, id)
	}

	// Ensure creator is in the list
	found := false
	for _, id := range memberIDs {
		if id == creatorID {
			found = true
			break
		}
	}
	if !found {
		memberIDs = append([]string{creatorID}, memberIDs...)
	}

	convID, err := rt.db.CreateConversation(creatorID, convType, name, photoData, memberIDs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": convID})
}