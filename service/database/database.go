/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rerikdev/WASAText/service/structures"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	Ping() error
	CheckUserExistence(username string) (bool, error)
	DoLogin(username, displayName, profilePicture string) (*structures.User, string, error)
	GetUserById(userId string) (*structures.User, error)
	SetMyPhotoById(userId, photoUrl string) error
	SetMyUserNameById(userId, newUsername string) error
	SearchUsers(query string) ([]*structures.User, error)
	// Conversazioni 1:1
	CreateConversation(user1, user2 int) (int64, error)
	// Messaggi
	SendMessage(conversationId, senderId int, content, mediaType string, isForwarded bool, replyToMessageId *int) ([]*structures.Message, error)
	GetMessages(conversationId int) ([]*structures.Message, error)
	// Restituisce tutte le conversazioni di un utente con anteprima ultimo messaggio
	GetUserConversations(userId int) ([]*structures.ConversationPreview, error)
	SetMessagesReceived(conversationId int, userId int) error
	SetMessagesRead(conversationId int, userId int) error
	DeleteMessage(conversationId int, messageId int, userId int) error
	GetMessageById(conversationId, messageId int) (*structures.Message, error)
	// Reazioni
	AddReaction(messageId int, userId int, emoji string) error
	RemoveReaction(messageId int, userId int) error
	GetReactions(messageId int) ([]*structures.Reaction, error)
	// Gruppi (usano la logica unificata delle conversazioni)
	AddToGroup(name string, photo string, usernames []string) (*structures.Conversation, error) // operationId: addToGroup
	ListGroups(userID int) ([]*structures.GroupPreview, error)                                  // operationId: listGroups
	LeaveGroup(groupID int, userID int) error
	SetGroupName(groupID int, newName string) error
	SetGroupPhoto(groupID int, photoUrl string) error
	AddMembersToGroup(groupID int, usernames []string) error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// It also creates all tables if the database is empty.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// SOLO PER SVILUPPO: elimina tutte le tabelle esistenti
	// if err := dropAllTables(db); err != nil {
	// return nil, fmt.Errorf("error dropping tables: %w", err)
	// }

	// Check if the main table exists. If not, the database is empty, and we need to create the structure
	var tableName string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		tables := []string{
			`CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                username TEXT NOT NULL UNIQUE,
                display_name TEXT NOT NULL,
                profile_picture TEXT
            );`,
			`CREATE TABLE IF NOT EXISTS conversations (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT,                -- NULL per chat 1:1, valorizzato per gruppi
                photo TEXT,               -- NULL per chat 1:1, valorizzato per gruppi
                is_group BOOLEAN NOT NULL DEFAULT 0
            );`,
			`CREATE TABLE IF NOT EXISTS conversation_members (
                conversation_id INTEGER NOT NULL,
                user_id INTEGER NOT NULL,
                PRIMARY KEY (conversation_id, user_id),
                FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
                FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
            );`,
			`CREATE TABLE IF NOT EXISTS messages (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                conversation_id INTEGER NOT NULL,
                sender_id INTEGER NOT NULL,
                content TEXT NOT NULL,
                is_forwarded BOOLEAN NOT NULL DEFAULT 0,
                media_type TEXT NOT NULL,
                status TEXT NOT NULL CHECK (status IN ('sent', 'received', 'read')),
                timestamp DATETIME NOT NULL,
                reply_to_message_id INTEGER DEFAULT NULL,
                FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
                FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
                FOREIGN KEY (reply_to_message_id) REFERENCES messages(id) ON DELETE SET NULL
            );`,
			`CREATE TABLE IF NOT EXISTS reactions (
                message_id INTEGER NOT NULL,
                user_id INTEGER NOT NULL,
                emoji TEXT NOT NULL,
                PRIMARY KEY (message_id, user_id),
                FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
                FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
            );`,
		}
		for _, stmt := range tables {
			if _, err := db.Exec(stmt); err != nil {
				return nil, fmt.Errorf("error creating database structure: %w", err)
			}
		}
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("error checking database structure: %w", err)
	}

	// Migration: Add reply_to_message_id column if it doesn't exist (for existing databases)
	var columnExists int
	err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('messages') WHERE name='reply_to_message_id';`).Scan(&columnExists)
	if err == nil && columnExists == 0 {
		_, err = db.Exec(`ALTER TABLE messages ADD COLUMN reply_to_message_id INTEGER DEFAULT NULL;`)
		if err != nil {
			return nil, fmt.Errorf("error adding reply_to_message_id column: %w", err)
		}
	}

	appdb := &appdbimpl{
		c: db,
	}

	return appdb, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
