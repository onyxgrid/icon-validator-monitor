package model

// User represents a user in the database
type User struct {
	// ID is the user's Telegram ID
	ID      int
	Email   *string
	Wallets []string
}