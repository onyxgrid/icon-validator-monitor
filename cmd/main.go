package main

import (
	"github.com/joho/godotenv"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
	"github.com/paulrouge/icon-validator-monitor/internal/icon"
	"github.com/paulrouge/icon-validator-monitor/internal/model"
	"github.com/paulrouge/icon-validator-monitor/internal/sender/mail"
	"github.com/paulrouge/icon-validator-monitor/internal/tg"
)

// todo:
// - Add correct logging troughout the code
// - Create a function that sends a weekly report to all users
// - get the correct omm vote power stuff

type MainService struct {
	TgBot *tg.TelegramBot
	// Senders is the list of senders that will be used to send the notifications
	Senders []model.Sender
	Icon *icon.Icon
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	
	// Create a new DB connection
	err = db.NewDB(); if err != nil {
		panic(err)
	}
	defer db.DBInstance.Close()

	err = db.DBInstance.Init(); if err != nil {
		panic(err)
	}

	// Create a new Icon client
	client, err := icon.NewIcon(); if err != nil {
		panic(err)
	}

	// Create a new Telegram bot
	tgBot, err := tg.NewBot(db.DBInstance, client); if err != nil {
		// this should be added to the log. (failed on parsing res into validotrinfo before...)	
		panic(err)
	}

	go tgBot.Init();
	
	// Create a new MainService
	service := NewMainService(db.DBInstance, tgBot, client)
	
	// Register the senders that will send the notifications
	service.registerSender(tgBot)

	// Create a gmail sender
	gmailSender, err := mail.NewMail(); if err != nil {
		panic(err)
	}

	// Register the gmail sender
	service.registerSender(gmailSender)

	select{}
}

func NewMainService(db *db.DB, tgBot *tg.TelegramBot, c *icon.Icon) *MainService {
	return &MainService{
		TgBot: tgBot,
		Icon: c,
	}
}

func (m *MainService) registerSender(sender model.Sender) {
	m.Senders = append(m.Senders, sender)
}

// to do; determin how to use this, sendmessage should take in receiver and message
func (m *MainService) sendMessage(to string, msg string) {
	// for _, sender := range m.Senders {
	// 	// sender.SendMessage(msg)
	// }
}