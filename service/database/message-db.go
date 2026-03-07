package database

import (
	"fmt"

	"github.com/rerikdev/WASAText/service/globaltime"
	"github.com/rerikdev/WASAText/service/structures"
)

// SendMessage inserisce un nuovo messaggio e restituisce la lista aggiornata dei messaggi della conversazione
func (db *appdbimpl) SendMessage(conversationId, senderId int, content, mediaType string, isForwarded bool, replyToMessageId *int) ([]*structures.Message, error) {
	// Controlla che la conversazione esista
	var exists int
	err := db.c.QueryRow(`SELECT COUNT(*) FROM conversations WHERE id = ?`, conversationId).Scan(&exists)
	if err != nil || exists == 0 {
		return nil, fmt.Errorf("conversazione non trovata")
	}
	// Controlla che il mittente sia membro della conversazione
	var count int
	err = db.c.QueryRow(`SELECT COUNT(*) FROM conversation_members WHERE conversation_id = ? AND user_id = ?`, conversationId, senderId).Scan(&count)
	if err != nil || count == 0 {
		return nil, fmt.Errorf("utente non autorizzato a inviare messaggi in questa conversazione")
	}

	// Se c'è un replyToMessageId, verifica che il messaggio esista nella stessa conversazione
	if replyToMessageId != nil {
		var replyExists int
		err = db.c.QueryRow(`SELECT COUNT(*) FROM messages WHERE id = ? AND conversation_id = ?`, *replyToMessageId, conversationId).Scan(&replyExists)
		if err != nil || replyExists == 0 {
			return nil, fmt.Errorf("messaggio di risposta non trovato nella conversazione")
		}
	}

	status := "received"

	_, err = db.c.Exec(
		`INSERT INTO messages (conversation_id, sender_id, content, is_forwarded, media_type, status, timestamp, reply_to_message_id)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		conversationId, senderId, content, isForwarded, mediaType, status, globaltime.Now().Format("2006-01-02 15:04:05"), replyToMessageId,
	)

	if err != nil {
		return nil, err
	}
	return db.GetMessages(conversationId)
}

// GetMessages restituisce tutti i messaggi di una conversazione, ordinati dal più vecchio al più recente
func (db *appdbimpl) GetMessages(conversationId int) ([]*structures.Message, error) {
	// Controlla che la conversazione esista
	var exists int
	err := db.c.QueryRow(`SELECT COUNT(*) FROM conversations WHERE id = ?`, conversationId).Scan(&exists)
	if err != nil || exists == 0 {
		return nil, fmt.Errorf("conversazione non trovata")
	}

	rows, err := db.c.Query(
		`SELECT m.id, m.conversation_id, m.sender_id, m.content, m.is_forwarded, m.media_type, m.status, m.timestamp, m.reply_to_message_id,
                u.username, u.display_name, u.profile_picture
         FROM messages m
         JOIN users u ON m.sender_id = u.id
         WHERE m.conversation_id = ?
         ORDER BY m.timestamp ASC`, conversationId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}

	var messages []*structures.Message
	messageMap := make(map[int]*structures.Message) // Per riferimenti veloci durante la costruzione delle reply

	// Prima passata: leggi tutti i messaggi
	for rows.Next() {
		var msg structures.Message
		var sender structures.User
		var replyToID *int

		if err := rows.Scan(
			&msg.ID, &msg.ConversationID, &sender.ID, &msg.Content, &msg.IsForwarded, &msg.MediaType, &msg.Status, &msg.Timestamp, &replyToID,
			&sender.Username, &sender.DisplayName, &sender.ProfilePicture,
		); err != nil {
			return nil, err
		}
		msg.Sender = sender
		msg.ReplyToMessageID = replyToID
		msg.Reactions = []*structures.Reaction{} // Inizializza slice vuoto

		messages = append(messages, &msg)
		messageMap[msg.ID] = &msg
	}

	// Seconda passata: popola i riferimenti alle reply e le reazioni
	for i := range messages {
		if messages[i].ReplyToMessageID != nil {
			if replyMsg, exists := messageMap[*messages[i].ReplyToMessageID]; exists {
				// Crea una versione semplificata senza reply annidate per evitare ricorsione infinita
				simplifiedReply := &structures.Message{
					ID:             replyMsg.ID,
					ConversationID: replyMsg.ConversationID,
					Content:        replyMsg.Content,
					MediaType:      replyMsg.MediaType,
					Sender:         replyMsg.Sender,
					Timestamp:      replyMsg.Timestamp,
					Status:         replyMsg.Status,
					IsForwarded:    replyMsg.IsForwarded,
				}
				messages[i].ReplyToMessage = simplifiedReply
			}
		}

		// Carica le reazioni per questo messaggio
		reactions, _ := db.GetReactions(messages[i].ID)
		messages[i].Reactions = reactions
	}

	return messages, nil
}

// Setta a "received" tutti i messaggi ricevuti dall'utente in una conversazione che sono ancora "sent"
func (db *appdbimpl) SetMessagesReceived(conversationId int, userId int) error {
	_, err := db.c.Exec(`
        UPDATE messages
        SET status = CASE WHEN status = 'sent' THEN 'received' ELSE status END
        WHERE conversation_id = ? AND sender_id != ?`, conversationId, userId)
	return err
}

// Setta a "read" tutti i messaggi ricevuti dall'utente in una conversazione che sono "received"
func (db *appdbimpl) SetMessagesRead(conversationId int, userId int) error {
	_, err := db.c.Exec(`
        UPDATE messages
        SET status = 'read'
        WHERE conversation_id = ? AND sender_id != ? AND status != 'read'`, conversationId, userId)
	return err
}

// DeleteMessage rimuove un messaggio da una conversazione se l'utente è il mittente
func (db *appdbimpl) DeleteMessage(conversationId, messageId, userId int) error {
	res, err := db.c.Exec(
		`DELETE FROM messages WHERE id = ? AND conversation_id = ? AND sender_id = ?`,
		messageId, conversationId, userId,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		return fmt.Errorf("not authorized or message not found")
	}
	return nil
}

// GetMessageById restituisce un messaggio specifico di una conversazione
func (db *appdbimpl) GetMessageById(conversationId, messageId int) (*structures.Message, error) {
	var msg structures.Message
	var sender structures.User
	var replyToID *int

	err := db.c.QueryRow(
		`SELECT m.id, m.conversation_id, m.sender_id, m.content, m.is_forwarded, m.media_type, m.status, m.timestamp, m.reply_to_message_id,
                u.username, u.display_name, u.profile_picture
         FROM messages m
         JOIN users u ON m.sender_id = u.id
         WHERE m.conversation_id = ? AND m.id = ?`, conversationId, messageId).Scan(
		&msg.ID, &msg.ConversationID, &sender.ID, &msg.Content, &msg.IsForwarded, &msg.MediaType, &msg.Status, &msg.Timestamp, &replyToID,
		&sender.Username, &sender.DisplayName, &sender.ProfilePicture,
	)

	if err != nil {
		return nil, fmt.Errorf("messaggio non trovato")
	}

	msg.Sender = sender
	msg.ReplyToMessageID = replyToID

	// Carica le reazioni
	reactions, _ := db.GetReactions(msg.ID)
	msg.Reactions = reactions

	// Se c'è un reply, caricalo (versione semplificata)
	if replyToID != nil {
		replyMsg, err := db.GetMessageById(conversationId, *replyToID)
		if err == nil {
			simplifiedReply := &structures.Message{
				ID:             replyMsg.ID,
				ConversationID: replyMsg.ConversationID,
				Content:        replyMsg.Content,
				MediaType:      replyMsg.MediaType,
				Sender:         replyMsg.Sender,
				Timestamp:      replyMsg.Timestamp,
				Status:         replyMsg.Status,
				IsForwarded:    replyMsg.IsForwarded,
			}
			msg.ReplyToMessage = simplifiedReply
		}
	}

	return &msg, nil
}
