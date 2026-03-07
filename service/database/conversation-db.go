package database

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/rerikdev/WASAText/service/structures"
)

// Crea una nuova conversazione 1:1 tra due utenti e restituisce l'id della conversazione
func (db *appdbimpl) CreateConversation(user1, user2 int) (int64, error) {
	if user1 == user2 {
		return 0, fmt.Errorf("non puoi creare una conversazione con te stesso")
	}

	// Controlla se esiste già una conversazione 1:1 tra questi due utenti
	var existingID int
	err := db.c.QueryRow(`
        SELECT c.id
        FROM conversations c
        JOIN conversation_members cm1 ON c.id = cm1.conversation_id AND cm1.user_id = ?
        JOIN conversation_members cm2 ON c.id = cm2.conversation_id AND cm2.user_id = ?
        WHERE c.is_group = 0
    `, user1, user2).Scan(&existingID)

	if err == nil && existingID != 0 {
		// Conversazione già esistente, restituisci l'ID esistente
		return int64(existingID), nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	// Crea la conversazione
	res, err := db.c.Exec(
		`INSERT INTO conversations (is_group) VALUES (0)`,
	)
	if err != nil {
		return 0, err
	}
	convID64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	convID := int(convID64)

	// Inserisci i due membri
	_, err = db.c.Exec(`INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?), (?, ?)`, convID, user1, convID, user2)
	if err != nil {
		return 0, err
	}

	return convID64, nil
}

// Restituisce tutte le conversazioni 1:1 dell'utente con anteprima ultimo messaggio
func (db *appdbimpl) GetUserConversations(userId int) ([]*structures.ConversationPreview, error) {
	// Prendi tutte le conversazioni dove l'utente è coinvolto
	rows, err := db.c.Query(`
        SELECT c.id
        FROM conversations c
        JOIN conversation_members cm ON c.id = cm.conversation_id
        WHERE cm.user_id = ? AND c.is_group = 0
    `, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var previews []*structures.ConversationPreview
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		// Trova l'altro utente
		var otherId int
		err = db.c.QueryRow(`
            SELECT user_id FROM conversation_members
            WHERE conversation_id = ? AND user_id != ?
        `, id, userId).Scan(&otherId)
		if err != nil {
			return nil, err
		}

		// Prendi info dell'altro utente
		var username, profilePicture string
		err = db.c.QueryRow(`SELECT username, profile_picture FROM users WHERE id = ?`, otherId).Scan(&username, &profilePicture)
		if err != nil {
			return nil, err
		}

		// Prendi ultimo messaggio (se esiste)
		var lastMsg, lastTime string
		_ = db.c.QueryRow(`
            SELECT content, timestamp FROM messages
            WHERE conversation_id = ?
            ORDER BY timestamp DESC LIMIT 1
        `, id).Scan(&lastMsg, &lastTime)

		previews = append(previews, &structures.ConversationPreview{
			ID:              id,
			OtherUserID:     otherId,
			Username:        username,
			ProfilePicture:  profilePicture,
			LastMessage:     lastMsg,
			LastMessageTime: lastTime,
		})
	}

	// Controlla errori del rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Ordina in ordine cronologico inverso (ultimo messaggio più recente)
	sort.Slice(previews, func(i, j int) bool {
		return previews[i].LastMessageTime > previews[j].LastMessageTime
	})

	return previews, nil
}
