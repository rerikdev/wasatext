package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rerikdev/WASAText/service/structures"
)

// Assicurati che _router abbia il campo db di tipo AppDatabase
// type _router struct { db database.AppDatabase }

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"message":"Method Not Allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req structures.SessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Name) < 3 || len(req.Name) > 16 {
		http.Error(w, `{"message":"Invalid request: name is required and must be 3-16 characters"}`, http.StatusBadRequest)
		return
	}

	user, action, err := rt.db.DoLogin(req.Name, req.DisplayName, req.ProfilePicture)
	if err != nil {
		if err.Error() == "registrazione già effettuata" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "Registrazione già effettuata"}); encErr != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		if err.Error() == "per la registrazione servono displayName e profilePicture" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "Utente non trovato, effettuare la registrazione"}); encErr != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		http.Error(w, `{"message":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	var msg string
	if action == "login" {
		msg = "Login effettuato"
	} else if action == "register" {
		msg = "Registrazione effettuata"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encErr := json.NewEncoder(w).Encode(struct {
		Message string           `json:"message"`
		User    *structures.User `json:"user"`
	}{
		Message: msg,
		User:    user,
	}); encErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
