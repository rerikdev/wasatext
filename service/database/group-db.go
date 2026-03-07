package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rerikdev/WASAText/service/structures"
)

// AddToGroup: crea un nuovo gruppo come conversazione con is_group = 1
func (db *appdbimpl) AddToGroup(name string, photo string, usernames []string) (*structures.Conversation, error) {
	tx, err := db.c.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// Log dell'errore se necessario
		}
	}()

	// Crea la conversazione di gruppo
	res, err := tx.Exec(`INSERT INTO conversations (name, photo, is_group) VALUES (?, ?, 1)`, name, photo)
	if err != nil {
		return nil, err
	}
	convID64, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	convID := int(convID64)

	members := make([]structures.User, 0) // oppure se conosci la dimensione: make([]structures.User, 0, expectedSize)
	for _, username := range usernames {
		var user structures.User
		err := db.c.QueryRow(`SELECT id, username, display_name, profile_picture FROM users WHERE username = ?`, username).
			Scan(&user.ID, &user.Username, &user.DisplayName, &user.ProfilePicture)
		if err != nil {
			return nil, fmt.Errorf("utente %s non trovato", username)
		}
		_, err = tx.Exec(`INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`, convID, user.ID)
		if err != nil {
			return nil, err
		}
		members = append(members, user)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &structures.Conversation{
		ID:      convID,
		Name:    name,
		Photo:   photo,
		IsGroup: true,
		Members: members,
	}, nil
}

// ListGroups: restituisce tutte le conversazioni di gruppo dell'utente
func (db *appdbimpl) ListGroups(userID int) ([]*structures.GroupPreview, error) {
	rows, err := db.c.Query(`
        SELECT c.id, COALESCE(c.name,''), COALESCE(c.photo,'')
        FROM conversations c
        JOIN conversation_members cm ON c.id = cm.conversation_id
        WHERE cm.user_id = ? AND c.is_group = 1
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*structures.GroupPreview
	for rows.Next() {
		var id int
		var name, photo string
		if err := rows.Scan(&id, &name, &photo); err != nil {
			return nil, err
		}

		// Prendi ultimo messaggio
		var lastMsg, lastTime string
		_ = db.c.QueryRow(`SELECT content, timestamp FROM messages WHERE conversation_id = ? ORDER BY timestamp DESC LIMIT 1`, id).Scan(&lastMsg, &lastTime)

		// Prendi membri del gruppo
		members, err := db.getConversationMembers(id)
		if err != nil {
			return nil, err
		}

		groups = append(groups, &structures.GroupPreview{
			ID:              id,
			Name:            name,
			Photo:           photo,
			Members:         members,
			LastMessage:     lastMsg,
			LastMessageTime: lastTime,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

// getConversationMembers: restituisce i membri di una conversazione
func (db *appdbimpl) getConversationMembers(convID int) ([]structures.User, error) {
	rows, err := db.c.Query(`
        SELECT u.id, u.username, u.display_name, COALESCE(u.profile_picture, '')
        FROM users u
        JOIN conversation_members cm ON u.id = cm.user_id
        WHERE cm.conversation_id = ?`, convID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []structures.User
	for rows.Next() {
		var user structures.User
		if err := rows.Scan(&user.ID, &user.Username, &user.DisplayName, &user.ProfilePicture); err != nil {
			return nil, err
		}
		members = append(members, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

// LeaveGroup: rimuove l'utente dalla conversazione di gruppo
func (db *appdbimpl) LeaveGroup(groupID int, userID int) error {
	var isGroup bool
	err := db.c.QueryRow(`SELECT is_group FROM conversations WHERE id = ?`, groupID).Scan(&isGroup)
	if err != nil {
		return fmt.Errorf("gruppo non trovato")
	}
	if !isGroup {
		return fmt.Errorf("non è un gruppo")
	}
	_, err = db.c.Exec(`DELETE FROM conversation_members WHERE conversation_id = ? AND user_id = ?`, groupID, userID)
	return err
}

// SetGroupName: aggiorna il nome del gruppo
func (db *appdbimpl) SetGroupName(groupID int, newName string) error {
	var isGroup bool
	err := db.c.QueryRow(`SELECT is_group FROM conversations WHERE id = ?`, groupID).Scan(&isGroup)
	if err != nil {
		return fmt.Errorf("gruppo non trovato")
	}
	if !isGroup {
		return fmt.Errorf("non è un gruppo")
	}
	_, err = db.c.Exec(`UPDATE conversations SET name = ? WHERE id = ?`, newName, groupID)
	return err
}

// SetGroupPhoto: aggiorna la foto del gruppo
func (db *appdbimpl) SetGroupPhoto(groupID int, photoUrl string) error {
	var isGroup bool
	err := db.c.QueryRow(`SELECT is_group FROM conversations WHERE id = ?`, groupID).Scan(&isGroup)
	if err != nil {
		return fmt.Errorf("gruppo non trovato")
	}
	if !isGroup {
		return fmt.Errorf("non è un gruppo")
	}
	_, err = db.c.Exec(`UPDATE conversations SET photo = ? WHERE id = ?`, photoUrl, groupID)
	return err
}

// AddMembersToGroup aggiunge utenti a un gruppo esistente
func (db *appdbimpl) AddMembersToGroup(groupID int, usernames []string) error {
	for _, username := range usernames {
		var userID int
		err := db.c.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&userID)
		if err != nil {
			continue // ignora utenti non trovati
		}
		_, _ = db.c.Exec(`INSERT OR IGNORE INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`, groupID, userID)
	}
	return nil
}
