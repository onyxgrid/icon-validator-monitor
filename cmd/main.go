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
// - Take out db from mainservice and make db a global variable, it's fine to have a global db connection
// - Add correct logging troughout the code
// - Create a function that sends a weekly report to all users

type MainService struct {
	db *db.DB
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
	db, err := db.NewDB(); if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Init(); if err != nil {
		panic(err)
	}

	// Create a new Icon client
	client, err := icon.NewIcon(); if err != nil {
		panic(err)
	}

	// Create a new Telegram bot
	tgBot, err := tg.NewBot(db, client); if err != nil {
		// this should be added to the log. (failed on parsing res into validotrinfo before...)	
		panic(err)
	}

	go tgBot.Init();
	
	// Create a new MainService
	service := NewMainService(db, tgBot, client)
	
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
		db: db,
		TgBot: tgBot,
		Icon: c,
	}
}

func (m *MainService) registerSender(sender model.Sender) {
	m.Senders = append(m.Senders, sender)
}

// to do; dettermin how to use this, sendmessage should take in receiver and message
func (m *MainService) sendMessage(to string, msg string) {
	// for _, sender := range m.Senders {
	// 	// sender.SendMessage(msg)
	// }
}