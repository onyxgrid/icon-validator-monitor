package db

import (
	"database/sql"
	"fmt"
	"slices"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/paulrouge/icon-validator-monitor/internal/config"
	"github.com/paulrouge/icon-validator-monitor/internal/model"
)

type DB struct {
	db *sql.DB
}

var DBInstance *DB

func NewDB() error {
	db, err := sql.Open("sqlite3", "./data/test_users.db")
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
			wallets TEXT DEFAULT '',
			alerts TEXT DEFAULT '',
			inactive BOOLEAN DEFAULT FALSE
		)
	`)
	// d.Migrate()
	return err
}

// Migrate migrates the database schema from the old schema to the new schema.
// use this function to migrate the database schema from the old schema to the new schema.
// run the function by calling db.Migrate()
func (d *DB) Migrate() {
	// prompt user to run the migration
	fmt.Println("Are you sure you want to run the migrate function? (y/n)")
	var input string
	fmt.Scanln(&input)
	if input != "y" {
		return
	}

	// add a new column to the users table called inactive
	// _, err := d.db.Exec(`
	// 	ALTER TABLE users
	// 	ADD COLUMN inactive BOOLEAN DEFAULT FALSE
	// `)

	// if err != nil {
	// 	fmt.Println("adding inactive", err)
	// 	return
	// }

	// create a user_new table
	// _, err := d.db.Exec(`
	// 	CREATE TABLE IF NOT EXISTS users_new (
	// 		id TEXT PRIMARY KEY,
	// 		email TEXT DEFAULT '',
	// 		wallets TEXT DEFAULT '',
	// 		alerts TEXT DEFAULT '',
	// 		inactive BOOLEAN DEFAULT FALSE
	// 	)
	// `)
	// if err != nil {
	// 	fmt.Println("creating", err)
	// 	return
	// }

	// copy the data from the old table to the new table
	// _, err = d.db.Exec(`
	// 	INSERT INTO users_new (id, email, wallets, alerts, inactive)
	// 	SELECT id, email, wallets, alerts, inactive
	// 	FROM users
	// `)
	// if err != nil {
	// 	fmt.Println("copying", err)
	// 	return
	// }

	// drop the old table
	// _, err = d.db.Exec(`
	// 	DROP TABLE users
	// `)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// rename the new table to the old table
	// _, err = d.db.Exec(`
	// 	ALTER TABLE users_new
	// 	RENAME TO users
	// `)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
}

func (d *DB) AddUser(id string) error {
	// check if the user already exists
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	_, err = d.db.Exec("INSERT INTO users (id) VALUES (?)", id)
	return err
}

func (d *DB) GetUser(id string) (model.User, error) {
	var user model.User
	var wallets sql.NullString
	var alerts sql.NullString

	err := d.db.QueryRow("SELECT email, wallets, alerts, inactive FROM users WHERE id = ?", id).Scan(&user.Email, &wallets, &alerts, &user.Inactive)
	if err != nil {
		return user, err
	}

	if wallets.Valid {
		user.Wallets = strings.Split(wallets.String, ",")
	} else {
		user.Wallets = []string{}
	}

	if alerts.Valid {
		user.Alerts = strings.Split(alerts.String, ",")
	} else {
		user.Alerts = []string{}
	}

	iid, err := strconv.Atoi(id)
	if err != nil {
		return user, err
	}
	user.ID = iid

	return user, nil
}

// SetUserEmail sets the email of a user given its id. If the user does not exist, it returns an error.
func (d *DB) SetUserEmail(id, email string) error {
	_, err := d.db.Exec("UPDATE users SET email = ? WHERE id = ?", email, id)
	if err != nil {
		fmt.Println("error setting email", err)
	}
	return err
}

// SetUserInactive sets the inactive status of a user given its id. If the user does not exist, it returns an error.
func (d *DB) SetUserInactive(id string, inactive bool) error {
	_, err := d.db.Exec("UPDATE users SET inactive = ? WHERE id = ?", inactive, id)
	if err != nil {
		fmt.Println("error setting inactive", err)
	}
	return err
}

// AddUserWallet adds a wallet to the user's wallets. If the user does not exist, it inserts a new row.
func (d *DB) AddUserWallet(id, wallet string) error {
	// wallets := d.GetUserWallets(id)
	u, err := d.GetUser(id)
	if err != nil {
		return err
	}

	if len(u.Wallets) == 0 {
		_, err := d.db.Exec("UPDATE users SET wallets = ? WHERE id = ?", wallet, id)
		return err
	} else {
		// add the wallet to the db
		u.Wallets = append(u.Wallets, wallet)
		// Join the wallets into a comma-separated string
		walletString := strings.Join(u.Wallets, ",")

		_, err := d.db.Exec("UPDATE users SET wallets = ? WHERE id = ?", walletString, id)
		return err
	}
}

// RemoveUserWallet removes a wallet from the user's wallets. If the user does not exist, it returns an error.
func (d *DB) RemoveUserWallet(id, wallet string) error {
	// wallets := d.GetUserWallets(id)
	u, err := d.GetUser(id)
	if err != nil {
		return err
	}

	// remove the wallet from the slice
	var newWallets []string
	for _, w := range u.Wallets {
		if w != wallet {
			newWallets = append(newWallets, w)
		}
	}

	// Join the wallets into a comma-separated string
	walletString := strings.Join(newWallets, ",")
	_, err = d.db.Exec("UPDATE users SET wallets = ? WHERE id = ?", walletString, id)
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

func (d *DB) AddAlert(id, alert string) error {
	if !slices.Contains(config.ALERTS, alert) {
		return fmt.Errorf("alert type %s is not valid", alert)
	}

	u, err := d.GetUser(id)
	if err != nil {
		return err
	}

	a := append(u.Alerts, alert)
	as := strings.Join(a, ",")

	_, err = d.db.Exec("UPDATE users SET alerts = ? WHERE id = ?", as, id)
	if err != nil {
		fmt.Println("error adding alert", err)
	}
	return err
}

func (d *DB) RemoveAlert(id, alert string) error {
	if !slices.Contains(config.ALERTS, alert) {
		return fmt.Errorf("alert type %s is not valid", alert)
	}

	// get the user's alerts
	var alerts string
	err := d.db.QueryRow("SELECT alerts FROM users WHERE id = ?", id).Scan(&alerts)
	if err != nil {
		return err
	}

	// split the alerts into a slice
	var a []string
	if alerts != "" {
		a = append(a, strings.Split(alerts, ",")...)
	}

	// remove the alert from the slice
	var newAlerts []string
	for _, a := range a {
		if a != alert {
			newAlerts = append(newAlerts, a)
		}
	}

	// a to []string
	var aString []string
	for _, alert := range newAlerts {
		aString = append(aString, string(alert))
	}

	// Join the alerts into a comma-separated string
	alertString := strings.Join(aString, ",")
	_, err = d.db.Exec("UPDATE users SET alerts = ? WHERE id = ?", alertString, id)
	return err
}

// GetUsersPerAlert returns a slice of user ids that have the given alert. If the alert type is not valid, it returns an error.
//
// Example of usage: GetUsersPerAlert("CPS")
func (d *DB) GetUsersPerAlert(alert string) ([]int64, error) {
	if !slices.Contains(config.ALERTS, alert) {
		return nil, fmt.Errorf("alert type %s is not valid", alert)
	}

	rows, err := d.db.Query("SELECT id FROM users WHERE alerts LIKE ?", "%"+alert+"%")
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
