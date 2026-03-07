package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// POST /conversations/:id/messages
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if !checkAuthorization(w, r) {
		return
	}
	conversationId, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "Conversazione non valida"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	userId, err := strconv.Atoi(r.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "Utente non autorizzato"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	var req struct {
		Content          string `json:"content"`
		MediaType        string `json:"mediaType"`
		IsForwarded      bool   `json:"isForwarded"`
		ReplyToMessageID *int   `json:"replyToMessageId"` // NEW: Optional reply reference
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "Contenuto mancante"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	if req.MediaType == "" {
		req.MediaType = "text"
	}

	// Validate reply message exists if provided
	if req.ReplyToMessageID != nil {
		_, err := rt.db.GetMessageById(conversationId, *req.ReplyToMessageID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "Messaggio di risposta non trovato"}); encErr != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
	}

	messages, err := rt.db.SendMessage(conversationId, userId, req.Content, req.MediaType, req.IsForwarded, req.ReplyToMessageID)
	if err != nil || len(messages) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"message": "Errore invio messaggio"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if encErr := json.NewEncoder(w).Encode(messages[len(messages)-1]); encErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GET /conversations/:id/messages
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if !checkAuthorization(w, r) {
		return
	}
	conversationId, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Conversazione non valida"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	userIdStr := r.Header.Get("Authorization")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Aggiorna a "received" i messaggi ricevuti da questo utente
	_ = rt.db.SetMessagesReceived(conversationId, userId)

	messages, err := rt.db.GetMessages(conversationId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Errore recupero messaggi"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if encErr := json.NewEncoder(w).Encode(messages); encErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// PUT /conversations/:id/messages/read
func (rt *_router) markMessagesRead(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if !checkAuthorization(w, r) {
		w.WriteHeader(http.StatusUnauthorized)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	conversationId, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Conversazione non valida"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	userIdStr := r.Header.Get("Authorization")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	err = rt.db.SetMessagesRead(conversationId, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Errore aggiornamento messaggi"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /conversations/:id/messages/:messageId
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if !checkAuthorization(w, r) {
		w.WriteHeader(http.StatusUnauthorized)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	conversationId, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Conversazione non valida"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	messageId, err := strconv.Atoi(ps.ByName("messageId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Messaggio non valido"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	userIdStr := r.Header.Get("Authorization")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	// Solo il mittente può eliminare il proprio messaggio
	err = rt.db.DeleteMessage(conversationId, messageId, userId)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Non autorizzato o errore eliminazione"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /conversations/:id/messages/:messageId/forward
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if !checkAuthorization(w, r) {
		w.WriteHeader(http.StatusUnauthorized)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	sourceConvId, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Conversazione non valida"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	messageId, err := strconv.Atoi(ps.ByName("messageId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Messaggio non valido"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	userIdStr := r.Header.Get("Authorization")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	var req struct {
		TargetConversationId int `json:"targetConversationId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TargetConversationId == 0 {
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "targetConversationId mancante o non valido"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Recupera il messaggio originale
	original, err := rt.db.GetMessageById(sourceConvId, messageId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Messaggio da inoltrare non trovato"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Inoltra il messaggio usando SendMessage (isForwarded: true, replyToMessageID: nil)
	messages, err := rt.db.SendMessage(req.TargetConversationId, userId, original.Content, original.MediaType, true, nil)
	if err != nil || len(messages) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Errore inoltro messaggio"}); encErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if encErr := json.NewEncoder(w).Encode(messages[len(messages)-1]); encErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
