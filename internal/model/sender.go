package model

type Sender interface {
	SendMessage(string, string) error
}