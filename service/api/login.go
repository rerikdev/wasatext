package api

import (
	"encoding/json"
	"net/http"
	"github.com/julienschmidt/httprouter"
)

// doLogin handles the POST /session requests 
func (rt *Router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. The structure for reading the JSON request body 
	var user struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 1. Check if user exists in the database
	id, err := rt.db.GetUserByName(user.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 2. If user doesn't exist -id is empty -,  create them
	if id == "" {
		id, err = rt.db.CreateUser(user.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// 3. Return the identifier as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	_ = json.NewEncoder(w).Encode(map[string]string{
		"identifier": id,
	})
}