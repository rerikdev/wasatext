package database

import (
	"database/sql"
	"fmt"

	"github.com/rerikdev/WASAText/service/structures"
)

// CheckUserExistence controlla se esiste un utente con lo username dato
func (db *appdbimpl) CheckUserExistence(username string) (bool, error) {
	var id int
	err := db.c.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// DoLogin controlla se l'utente esiste, se sì restituisce i dati e "login",
// se no lo crea e restituisce i dati e "register".
// Se si tenta di registrare un utente già esistente, restituisce errore.
func (db *appdbimpl) DoLogin(username, displayName, profilePicture string) (*structures.User, string, error) {
	exists, err := db.CheckUserExistence(username)
	if err != nil {
		return nil, "", err
	}

	if exists {
		// Se displayName o profilePicture sono forniti, significa che si sta tentando una doppia registrazione
		if displayName != "" || profilePicture != "" {
			return nil, "", fmt.Errorf("registrazione già effettuata")
		}
		// Recupera i dati dell'utente
		var user structures.User
		err := db.c.QueryRow(
			`SELECT id, username, display_name, profile_picture FROM users WHERE username = ?`, username,
		).Scan(&user.ID, &user.Username, &user.DisplayName, &user.ProfilePicture)
		if err != nil {
			return nil, "", err
		}
		return &user, "login", nil
	}

	// Se manca displayName o profilePicture, non si può registrare
	if displayName == "" || profilePicture == "" {
		return nil, "", fmt.Errorf("per la registrazione servono displayName e profilePicture")
	}

	// Crea nuovo utente
	res, err := db.c.Exec(
		`INSERT INTO users (username, display_name, profile_picture) VALUES (?, ?, ?)`,
		username, displayName, profilePicture,
	)
	if err != nil {
		return nil, "", err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, "", err
	}
	return &structures.User{
		ID:             int(id),
		Username:       username,
		DisplayName:    displayName,
		ProfilePicture: profilePicture,
	}, "register", nil
}

// GetUserById restituisce i dati di un utente dato il suo ID
func (db *appdbimpl) GetUserById(userId string) (*structures.User, error) {
	var user structures.User
	err := db.c.QueryRow(
		`SELECT id, username, display_name, profile_picture FROM users WHERE id = ?`, userId,
	).Scan(&user.ID, &user.Username, &user.DisplayName, &user.ProfilePicture)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SetMyUserNameById aggiorna lo username dell'utente dato il suo ID
func (db *appdbimpl) SetMyUserNameById(userId, newUsername string) error {
	// Usa CheckUserExistence per vedere se lo username è già preso da altri
	exists, err := db.CheckUserExistence(newUsername)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("username già in uso")
	}
	_, err = db.c.Exec(`UPDATE users SET username = ? WHERE id = ?`, newUsername, userId)
	return err
}

// SetMyPhotoById aggiorna la foto profilo dell'utente dato il suo ID
func (db *appdbimpl) SetMyPhotoById(userId, photoUrl string) error {
	_, err := db.c.Exec(`UPDATE users SET profile_picture = ? WHERE id = ?`, photoUrl, userId)
	return err
}

// SearchUsers restituisce una lista di utenti il cui username contiene la query
func (db *appdbimpl) SearchUsers(query string) ([]*structures.User, error) {
	rows, err := db.c.Query(
		`SELECT id, username, display_name, profile_picture 
         FROM users 
         WHERE username LIKE ? OR display_name LIKE ? 
         LIMIT 10`,
		"%"+query+"%", "%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*structures.User
	for rows.Next() {
		var user structures.User
		if err := rows.Scan(&user.ID, &user.Username, &user.DisplayName, &user.ProfilePicture); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	// AGGIUNGI CONTROLLO rows.Err():
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
