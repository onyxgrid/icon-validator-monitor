package model

type Sender interface {
	// Send message to receiver
	SendMessage(string, string) error
	// Send alert to receiver
	SendAlert(string, string, string) error
	// Get receiver for sender, parameter is the uid of the sender
	GetReceiver(string) string
}