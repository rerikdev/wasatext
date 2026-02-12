package api

import (
    "encoding/json"
    "io" // Added for binary reading
    "net/http"
    "strings"
    "github.com/julienschmidt/httprouter"
)

func (rt *Router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    id := rt.extractIdentifier(r)
    if id == "" {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    var body struct {
        Name string `json:"name"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    err := rt.db.UpdateUserName(id, body.Name)
    if err != nil {
        if strings.Contains(err.Error(), "UNIQUE constraint failed") {
            w.WriteHeader(http.StatusConflict)
            return
        }
        w.WriteHeader(http.StatusInternalServerError)
        return  
    }
    w.WriteHeader(http.StatusNoContent)
}

func (rt *Router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    id := rt.extractIdentifier(r)
    if id == "" {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    // Read the raw binary from the request body
    imgData, err := io.ReadAll(r.Body)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    // Pass the []byte to the database
    err = rt.db.UpdateUserPhoto(id, imgData)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}