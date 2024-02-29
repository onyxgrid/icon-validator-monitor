package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

var DBInstance *DB

func NewDB() error {
	db, err := sql.Open("sqlite3", "./data/users.db")
	if err != nil {
		return err
	}

	DBInstance = &DB{db: db}
	return nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Init() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT,
			wallets TEXT
		)
	`)
	return err
}

func (d *DB) AddUser(id string) error {
	// check if the user already exists
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		fmt.Println("user already exists")
		return nil
	}

	_, err = d.db.Exec("INSERT INTO users (id) VALUES (?)", id)
	return err
}

// GetUserEmail returns the email of a user given its id. If the user does not exist, it returns an empty string.
func (d *DB) GetUserEmail(id string) string {
	var email string
	err := d.db.QueryRow("SELECT email FROM users WHERE id = ?", id).Scan(&email)
	if err != nil{
		return ""
	}
	return email
}

// SetUserEmail sets the email of a user given its id. If the user does not exist, it returns an error.
func (d *DB) SetUserEmail(id, email string) error {
	_, err := d.db.Exec("UPDATE users SET email = ? WHERE id = ?", email, id)
	if err != nil {
		fmt.Println("error setting email", err)
	}
	return err
}

// GetUserWallets returns a slice of wallets of a user given its id. If the user does not exist, or has no wallets, it returns an empty slice.
func (d *DB) GetUserWallets(id string) ([]string) {
	var wallets string
	err := d.db.QueryRow("SELECT wallets FROM users WHERE id = ?", id).Scan(&wallets)
	if err != nil {
		return []string{}
	}

	// split the wallets into a slice
	var w []string
	if wallets != "" {
		w = strings.Split(wallets, ",")
	}

	return w
}

// AddUserWallet adds a wallet to the user's wallets. If the user does not exist, it inserts a new row.
func (d *DB) AddUserWallet(id, wallet string) error {
	wallets := d.GetUserWallets(id)

	if len(wallets) == 0 {
		_, err := d.db.Exec("UPDATE users SET wallets = ? WHERE id = ?", wallet, id)
		return err
	} else {
		// add the wallet to the db
		wallets = append(wallets, wallet)
        // Join the wallets into a comma-separated string
        walletString := strings.Join(wallets, ",")

		_, err := d.db.Exec("UPDATE users SET wallets = ? WHERE id = ?", walletString, id)
		return err
	}
}

// RemoveUserWallet removes a wallet from the user's wallets. If the user does not exist, it returns an error.
func (d *DB) RemoveUserWallet(id, wallet string) error {
	wallets := d.GetUserWallets(id)

	// remove the wallet from the slice
	var newWallets []string
	for _, w := range wallets {
		if w != wallet {
			newWallets = append(newWallets, w)
		}
	}

	// Join the wallets into a comma-separated string
	walletString := strings.Join(newWallets, ",")
	_, err := d.db.Exec("UPDATE users SET wallets = ? WHERE id = ?", walletString, id)
	return err
}

// RemoveAllWalletsUser removes all wallets from the user's wallets. If the user does not exist, it returns an error.
func (d *DB) RemoveAllWalletsUser(id string) error {
	_, err := d.db.Exec("UPDATE users SET wallets = ? WHERE id = ?", "", id)
	return err
}

// GetAllUsers returns a slice of all users in the database.
func (d *DB) GetAllUserIDs() ([]int64, error) {
	rows, err := d.db.Query("SELECT id FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			continue
		}
		users = append(users, id)
	}
	return users, nil
}

