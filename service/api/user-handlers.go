package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if !checkAuthorization(w, r) {
		return
	}
	userId := ps.ByName("userId")
	user, err := rt.db.GetUserById(userId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "Utente non trovato"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if encErr := json.NewEncoder(w).Encode(user); encErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// PATCH /users/:username/photo per cambiare foto profilo
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if !checkAuthorization(w, r) {
		return
	}
	userId := ps.ByName("userId")
	var req struct {
		PhotoUrl string `json:"photoUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PhotoUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "URL non valido"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	if err := rt.db.SetMyPhotoById(userId, req.PhotoUrl); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "Errore aggiornamento immagine"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	user, _ := rt.db.GetUserById(userId)
	w.Header().Set("Content-Type", "application/json")
	if encErr := json.NewEncoder(w).Encode(user); encErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// PATCH /users/:username per cambiare username
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if !checkAuthorization(w, r) {
		return
	}
	userId := ps.ByName("userId")
	var req struct {
		NewName string `json:"newName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NewName == "" {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "Username non valido"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	if err := rt.db.SetMyUserNameById(userId, req.NewName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "Useraname già in uso"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	user, _ := rt.db.GetUserById(userId)
	w.Header().Set("Content-Type", "application/json")
	if encErr := json.NewEncoder(w).Encode(user); encErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GET /users/search?q=...
func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !checkAuthorization(w, r) {
		return
	}
	query := r.URL.Query().Get("q")
	if len(query) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "Query troppo corta"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	users, err := rt.db.SearchUsers(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "Errore ricerca utenti"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if encErr := json.NewEncoder(w).Encode(users); encErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
