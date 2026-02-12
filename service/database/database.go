package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// AppDatabase interface
type AppDatabase interface {
	// Users
	CreateUser(name string) (string, error)
	GetUserByName(name string) (string, error)
	GetUsers(nameQuery string) ([]string, error)
	UpdateUserName(id string, name string) error
	GetUser(id string) (string, string, []byte, error)
	UpdateUserPhoto(id string, data []byte) error

	// Conversations (direct + group)
	CreateConversation(creatorID string, convType string, name string, photo []byte, memberIDs []string) (string, error)
	GetConversations(userID string) ([]map[string]interface{}, error)
	GetConversation(convID string, userID string) ([]map[string]interface{}, error)
	UpdateConversation(convID string, name string, photo []byte) error
	DeleteConversation(convID string, userID string) error
	GetConversationType(convID string) (string, error)
	GetDirectConversationID(user1, user2 string) (string, error)

	// Participants
	AddParticipant(convID string, userID string) error
	RemoveParticipant(convID string, userID string) error
	GetParticipants(convID string) ([]map[string]interface{}, error)

	// Messages
	SendMessage(convID string, senderID string, content string, image []byte, replyTo *string) error
	DeleteMessage(msgID string, senderID string) error
	MarkMessagesAsRead(convID string, userID string, messageID *string) error

	// Reactions
	AddReaction(messageID string, userID string, emoticon string) error
	RemoveReaction(messageID string, userID string, emoticon string) error

    GetMessageByID(messageID string) (content string, image []byte, err error)
    IsParticipant(convID string, userID string) (bool, error)
    GetConversationCreator(convID string) (string, error)

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New creates a new database instance and runs all migrations
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}

	// ------------------------------------------------------------
	// 1. Base tables (users, messages, reactions)
	// ------------------------------------------------------------
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		photo BLOB
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_id TEXT NOT NULL,
		receiver_id TEXT,                     -- NULL for group messages
		conversation_id INTEGER NOT NULL,     -- group chat support
		content TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		status TEXT DEFAULT 'sent',
		image BLOB,
		reply_to INTEGER,
		FOREIGN KEY(sender_id) REFERENCES users(id),
		FOREIGN KEY(receiver_id) REFERENCES users(id),
		FOREIGN KEY(conversation_id) REFERENCES conversations(id),
		FOREIGN KEY(reply_to) REFERENCES messages(id)
	);

	CREATE TABLE IF NOT EXISTS reactions (
		message_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		emoticon TEXT NOT NULL,
		PRIMARY KEY (message_id, user_id),
		FOREIGN KEY(message_id) REFERENCES messages(id),
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	`)
	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 2. New tables for group chats
	// ------------------------------------------------------------
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS conversations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,                 -- 'direct' or 'group'
		name TEXT,                         -- group name, NULL for direct
		photo BLOB,                        -- group photo, NULL for direct
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_by INTEGER NOT NULL,
		FOREIGN KEY(created_by) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS participants (
		conversation_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		left_at DATETIME,                  -- NULL if still member
		PRIMARY KEY (conversation_id, user_id),
		FOREIGN KEY(conversation_id) REFERENCES conversations(id),
		FOREIGN KEY(user_id) REFERENCES users(id)
	);

	CREATE INDEX idx_participants_user ON participants(user_id);
	CREATE INDEX idx_messages_conversation ON messages(conversation_id);
	`)
	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 3. Add missing columns to messages (idempotent)
	// ------------------------------------------------------------
	migrations := []string{
		`ALTER TABLE messages ADD COLUMN conversation_id INTEGER`,
		`ALTER TABLE messages ADD COLUMN receiver_id TEXT`,
		`ALTER TABLE messages ADD COLUMN image BLOB`,
		`ALTER TABLE messages ADD COLUMN reply_to INTEGER`,
	}
	for _, stmt := range migrations {
		_, err = db.Exec(stmt)
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return nil, err
		}
	}

	// ------------------------------------------------------------
	// 4. Existing direct messages will be migrated lazily
	//    in GetConversations() / ensureDirectConversations()
	// ------------------------------------------------------------
	return &appdbimpl{c: db}, nil
}

// ----------------------------------------------------------------------
// Conversations
// ----------------------------------------------------------------------

// CreateConversation creates a new conversation (direct or group)
func (db *appdbimpl) CreateConversation(creatorID string, convType string, name string, photo []byte, memberIDs []string) (string, error) {
	tx, err := db.c.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		INSERT INTO conversations (type, name, photo, created_by)
		VALUES (?, ?, ?, ?)`,
		convType, name, photo, creatorID)
	if err != nil {
		return "", err
	}
	convID, err := res.LastInsertId()
	if err != nil {
		return "", err
	}
	convIDStr := fmt.Sprint(convID)

	// Add creator as participant
	err = db.addParticipantTx(tx, convIDStr, creatorID)
	if err != nil {
		return "", err
	}

	// Add other members
	for _, uid := range memberIDs {
		if uid == creatorID {
			continue
		}
		err = db.addParticipantTx(tx, convIDStr, uid)
		if err != nil {
			return "", err
		}
	}

	err = tx.Commit()
	return convIDStr, err
}

// addParticipantTx adds a user to a conversation (inside a transaction)
func (db *appdbimpl) addParticipantTx(tx *sql.Tx, convID string, userID string) error {
	_, err := tx.Exec(`
		INSERT OR IGNORE INTO participants (conversation_id, user_id, joined_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)`,
		convID, userID)
	return err
}

// AddParticipant public method
func (db *appdbimpl) AddParticipant(convID string, userID string) error {
	_, err := db.c.Exec(`
		INSERT OR IGNORE INTO participants (conversation_id, user_id, joined_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)`,
		convID, userID)
	return err
}

// RemoveParticipant marks a user as left (soft delete)
func (db *appdbimpl) RemoveParticipant(convID string, userID string) error {
	_, err := db.c.Exec(`
		UPDATE participants SET left_at = CURRENT_TIMESTAMP
		WHERE conversation_id = ? AND user_id = ? AND left_at IS NULL`,
		convID, userID)
	return err
}

// GetParticipants returns all active participants of a conversation
func (db *appdbimpl) GetParticipants(convID string) ([]map[string]interface{}, error) {
	rows, err := db.c.Query(`
		SELECT u.id, u.name, u.photo
		FROM participants p
		JOIN users u ON p.user_id = u.id
		WHERE p.conversation_id = ? AND p.left_at IS NULL
		ORDER BY p.joined_at ASC`, convID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []map[string]interface{}
	for rows.Next() {
		var id, name string
		var photo []byte
		if err := rows.Scan(&id, &name, &photo); err != nil {
			return nil, err
		}
		participants = append(participants, map[string]interface{}{
			"id":    id,
			"name":  name,
			"photo": photo,
		})
	}
	return participants, rows.Err()
}

// GetConversations returns all conversations the user is part of (active)
func (db *appdbimpl) GetConversations(userID string) ([]map[string]interface{}, error) {
	// First, ensure any direct conversation that should exist does exist (lazy creation)
	err := db.ensureDirectConversations(userID)
	if err != nil {
		return nil, err
	}

	rows, err := db.c.Query(`
		SELECT 
			c.id, c.type, c.name, c.photo,
			(SELECT COUNT(*) FROM participants p2 WHERE p2.conversation_id = c.id AND p2.left_at IS NULL) as member_count,
			(SELECT m.content FROM messages m WHERE m.conversation_id = c.id ORDER BY m.timestamp DESC LIMIT 1) as last_message,
			(SELECT m.timestamp FROM messages m WHERE m.conversation_id = c.id ORDER BY m.timestamp DESC LIMIT 1) as last_timestamp,
			(SELECT u.name FROM users u WHERE u.id = c.created_by) as created_by_name
		FROM participants p
		JOIN conversations c ON p.conversation_id = c.id
		WHERE p.user_id = ? AND p.left_at IS NULL
		ORDER BY last_timestamp DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []map[string]interface{}
	for rows.Next() {
		var id, convType string
		var name, photo, lastMessage, lastTimestamp, createdByName sql.NullString
		var memberCount int
		err := rows.Scan(&id, &convType, &name, &photo, &memberCount, &lastMessage, &lastTimestamp, &createdByName)
		if err != nil {
			return nil, err
		}

		conv := map[string]interface{}{
			"id":          id,
			"type":        convType,
			"memberCount": memberCount,
		}
		if convType == "group" {
			conv["name"] = name.String
			conv["photo"] = photo.String // base64 later
		} else {
			// For direct chats, the name is the other participant's name
			otherName, err := db.getOtherParticipantName(id, userID)
			if err == nil {
				conv["name"] = otherName
			} else {
				conv["name"] = "Unknown User"
			}
		}
		if lastMessage.Valid {
			conv["lastMessage"] = map[string]interface{}{
				"preview":   lastMessage.String,
				"timestamp": lastTimestamp.String,
			}
		}
		conversations = append(conversations, conv)
	}
	return conversations, rows.Err()
}

// ensureDirectConversations creates missing direct conversations for a user
func (db *appdbimpl) ensureDirectConversations(userID string) error {
	// Find all distinct contacts the user has exchanged messages with
	rows, err := db.c.Query(`
		SELECT DISTINCT 
			CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END as contact_id
		FROM messages 
		WHERE (sender_id = ? OR receiver_id = ?) AND conversation_id IS NULL`,
		userID, userID, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var contactID string
		if err := rows.Scan(&contactID); err != nil {
			return err
		}
		// Check if a direct conversation already exists between these two
		var exists int
		err = db.c.QueryRow(`
			SELECT 1 FROM conversations c
			JOIN participants p1 ON c.id = p1.conversation_id
			JOIN participants p2 ON c.id = p2.conversation_id
			WHERE c.type = 'direct'
				AND p1.user_id = ? AND p2.user_id = ?
				AND p1.left_at IS NULL AND p2.left_at IS NULL`, userID, contactID).Scan(&exists)
		if err == sql.ErrNoRows {
			// Create conversation
			_, err = db.CreateConversation(userID, "direct", "", nil, []string{contactID})
			if err != nil {
				return err
			}
			// Update existing messages to point to this conversation
			convID, _ := db.GetDirectConversationID(userID, contactID) // ✅ FIXED: uppercase
			if convID != "" {
				_, err = db.c.Exec(`
					UPDATE messages SET conversation_id = ?
					WHERE (sender_id = ? AND receiver_id = ?)
					   OR (sender_id = ? AND receiver_id = ?)`,
					convID, userID, contactID, contactID, userID)
				if err != nil {
					return err
				}
			}
		} else if err != nil {
			return err
		}
	}
	return rows.Err()
}

// GetDirectConversationID returns the conversation ID for a direct chat between two users
func (db *appdbimpl) GetDirectConversationID(user1, user2 string) (string, error) {
	var convID string
	err := db.c.QueryRow(`
		SELECT c.id FROM conversations c
		JOIN participants p1 ON c.id = p1.conversation_id
		JOIN participants p2 ON c.id = p2.conversation_id
		WHERE c.type = 'direct'
			AND p1.user_id = ? AND p2.user_id = ?
			AND p1.left_at IS NULL AND p2.left_at IS NULL`,
		user1, user2).Scan(&convID)
	if err != nil {
		return "", err
	}
	return convID, nil
}

// getOtherParticipantName returns the name of the other user in a direct conversation
func (db *appdbimpl) getOtherParticipantName(convID string, myID string) (string, error) {
	var name string
	err := db.c.QueryRow(`
		SELECT u.name FROM participants p
		JOIN users u ON p.user_id = u.id
		WHERE p.conversation_id = ? AND p.user_id != ? AND p.left_at IS NULL`,
		convID, myID).Scan(&name)
	return name, err
}

// GetConversation returns all messages in a conversation (for the given user, who must be a participant)
func (db *appdbimpl) GetConversation(convID string, userID string) ([]map[string]interface{}, error) {
	// Verify user is a participant
	var active int
	err := db.c.QueryRow(`
		SELECT 1 FROM participants 
		WHERE conversation_id = ? AND user_id = ? AND left_at IS NULL`,
		convID, userID).Scan(&active)
	if err != nil {
		return nil, errors.New("access denied or conversation not found")
	}

	rows, err := db.c.Query(`
		SELECT 
			m.id, m.sender_id, m.content, m.timestamp, m.status, m.image,
			m.reply_to,
			orig.content as reply_content,
			orig.image as reply_image
		FROM messages m
		LEFT JOIN messages orig ON m.reply_to = orig.id
		WHERE m.conversation_id = ?
		ORDER BY m.timestamp ASC`, convID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var id, senderID, content, timestamp, status string
		var image []byte
		var replyTo sql.NullString
		var replyContent sql.NullString
		var replyImage []byte

		err := rows.Scan(&id, &senderID, &content, &timestamp, &status, &image,
			&replyTo, &replyContent, &replyImage)
		if err != nil {
			return nil, err
		}

		// Reactions with username
		reactionRows, err := db.c.Query(`
			SELECT r.user_id, r.emoticon, COALESCE(u.name, 'Deleted User')
			FROM reactions r
			LEFT JOIN users u ON r.user_id = u.id
			WHERE r.message_id = ?`, id)
		if err != nil {
			return nil, err
		}
		defer reactionRows.Close()

		var reactions []map[string]interface{}
		for reactionRows.Next() {
			var uid, emoticon, userName string
			if err := reactionRows.Scan(&uid, &emoticon, &userName); err != nil {
				return nil, err
			}
			reactions = append(reactions, map[string]interface{}{
				"user":     uid,
				"emoticon": emoticon,
				"userName": userName,
			})
		}
		if err = reactionRows.Err(); err != nil {
			return nil, err
		}

		// Sender object
		var senderName string
		_ = db.c.QueryRow(`SELECT name FROM users WHERE id = ?`, senderID).Scan(&senderName)
		sender := map[string]interface{}{
			"id":   senderID,
			"name": senderName,
		}

		msg := map[string]interface{}{
			"id":        id,
			"sender":    sender,
			"content":   content,
			"timestamp": timestamp,
			"status":    status,
			"reactions": reactions,
			"image":     image,
		}

		if replyTo.Valid {
			msg["reply"] = map[string]interface{}{
				"id":       replyTo.String,
				"content":  replyContent.String,
				"hasImage": replyImage != nil && len(replyImage) > 0,
			}
		}

		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

// UpdateConversation updates group name and/or photo
func (db *appdbimpl) UpdateConversation(convID string, name string, photo []byte) error {
	_, err := db.c.Exec(`
		UPDATE conversations SET name = ?, photo = ? WHERE id = ?`,
		name, photo, convID)
	return err
}

// DeleteConversation (soft delete for groups? For now, hard delete messages and conversation)
func (db *appdbimpl) DeleteConversation(convID string, userID string) error {
	// Check if user is the creator or has permission? For now, just delete if participant
	// This is a simplified version; real implementation should check permissions
	_, err := db.c.Exec(`DELETE FROM messages WHERE conversation_id = ?`, convID)
	if err != nil {
		return err
	}
	_, err = db.c.Exec(`DELETE FROM participants WHERE conversation_id = ?`, convID)
	if err != nil {
		return err
	}
	_, err = db.c.Exec(`DELETE FROM conversations WHERE id = ?`, convID)
	return err
}

// ----------------------------------------------------------------------
// Messages
// ----------------------------------------------------------------------

// SendMessage inserts a new message into a conversation
func (db *appdbimpl) SendMessage(convID string, senderID string, content string, image []byte, replyTo *string) error {
	// Verify sender is a participant
	var active int
	err := db.c.QueryRow(`
		SELECT 1 FROM participants 
		WHERE conversation_id = ? AND user_id = ? AND left_at IS NULL`,
		convID, senderID).Scan(&active)
	if err != nil {
		return errors.New("user not in conversation")
	}

	_, err = db.c.Exec(`
		INSERT INTO messages (conversation_id, sender_id, content, image, reply_to, status)
		VALUES (?, ?, ?, ?, ?, 'sent')`,
		convID, senderID, content, image, replyTo)
	return err
}

// DeleteMessage (unchanged, but now checks conversation as well? We'll keep original)
func (db *appdbimpl) DeleteMessage(msgID string, senderID string) error {
	res, err := db.c.Exec(`
		DELETE FROM messages 
		WHERE id = ? AND sender_id = ?`, msgID, senderID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.New("message not found or you are not the sender")
	}
	return nil
}

// MarkMessagesAsRead marks all messages in a conversation up to a certain point as read for a user
func (db *appdbimpl) MarkMessagesAsRead(convID string, userID string, messageID *string) error {
	// For simplicity, mark all messages in the conversation sent by others as read
	// In a full implementation, we'd store per-user read receipts.
	// For now, we just update the status of messages to 'read' for the receiver.
	// This works for direct chats; for groups we need a different approach.
	// We'll skip group read receipts for now and implement later.
	if messageID == nil {
		// Mark all as read
		_, err := db.c.Exec(`
			UPDATE messages SET status = 'read'
			WHERE conversation_id = ? AND sender_id != ? AND status != 'read'`,
			convID, userID)
		return err
	}
	// Mark up to specific message
	_, err := db.c.Exec(`
		UPDATE messages SET status = 'read'
		WHERE conversation_id = ? AND sender_id != ? AND id <= ? AND status != 'read'`,
		convID, userID, messageID)
	return err
}

// ----------------------------------------------------------------------
// Reactions
// ----------------------------------------------------------------------
func (db *appdbimpl) AddReaction(messageID string, userID string, emoticon string) error {
	_, err := db.c.Exec(`
		INSERT OR REPLACE INTO reactions (message_id, user_id, emoticon) 
		VALUES (?, ?, ?)`, messageID, userID, emoticon)
	return err
}

func (db *appdbimpl) RemoveReaction(messageID string, userID string, emoticon string) error {
	_, err := db.c.Exec(`
		DELETE FROM reactions 
		WHERE message_id = ? AND user_id = ? AND emoticon = ?`,
		messageID, userID, emoticon)
	return err
}

// ----------------------------------------------------------------------
// Users
// ----------------------------------------------------------------------
func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

func (db *appdbimpl) CreateUser(name string) (string, error) {
	res, err := db.c.Exec("INSERT INTO users (name) VALUES (?)", name)
	if err != nil {
		return "", err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return "", err
	}
	return fmt.Sprint(lastID), nil
}

func (db *appdbimpl) GetUserByName(name string) (string, error) {
	var id int
	err := db.c.QueryRow("SELECT id FROM users WHERE name = ?", name).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return fmt.Sprint(id), nil
}

func (db *appdbimpl) GetUsers(nameQuery string) ([]string, error) {
	rows, err := db.c.Query("SELECT name FROM users WHERE name LIKE ?", "%"+nameQuery+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		users = append(users, name)
	}
	return users, rows.Err()
}

func (db *appdbimpl) UpdateUserName(id string, name string) error {
	_, err := db.c.Exec("UPDATE users SET name = ? WHERE id = ?", name, id)
	return err
}

func (db *appdbimpl) GetUser(id string) (string, string, []byte, error) {
	var dbID, name string
	var photo []byte
	err := db.c.QueryRow("SELECT id, name, photo FROM users WHERE id = ?", id).Scan(&dbID, &name, &photo)
	if err != nil {
		return "", "", nil, err
	}
	return dbID, name, photo, nil
}

func (db *appdbimpl) UpdateUserPhoto(id string, data []byte) error {
	_, err := db.c.Exec("UPDATE users SET photo = ? WHERE id = ?", data, id)
	return err
}

// GetConversationType returns the type ('direct' or 'group') of a conversation.
func (db *appdbimpl) GetConversationType(convID string) (string, error) {
	var convType string
	err := db.c.QueryRow(`SELECT type FROM conversations WHERE id = ?`, convID).Scan(&convType)
	return convType, err
}

// GetMessageByID returns the content and image of a message.
func (db *appdbimpl) GetMessageByID(messageID string) (string, []byte, error) {
	var content string
	var image []byte
	err := db.c.QueryRow(`SELECT content, image FROM messages WHERE id = ?`, messageID).Scan(&content, &image)
	return content, image, err
}

// IsParticipant checks if a user is currently an active participant in a conversation.
func (db *appdbimpl) IsParticipant(convID string, userID string) (bool, error) {
	var count int
	err := db.c.QueryRow(`
		SELECT COUNT(*) FROM participants
		WHERE conversation_id = ? AND user_id = ? AND left_at IS NULL
	`, convID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetConversationCreator returns the user ID who created the conversation.
func (db *appdbimpl) GetConversationCreator(convID string) (string, error) {
	var creatorID string
	err := db.c.QueryRow(`SELECT created_by FROM conversations WHERE id = ?`, convID).Scan(&creatorID)
	return creatorID, err
}

