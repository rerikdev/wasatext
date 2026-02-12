package api

import (
    "encoding/json"
    "net/http"
    "github.com/julienschmidt/httprouter"
)

func (rt *Router) getUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    // Get the 'name' search parameter from the URL: /users?name
    query := r.URL.Query().Get("name")

    users, err := rt.db.GetUsers(query)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(users)
}
