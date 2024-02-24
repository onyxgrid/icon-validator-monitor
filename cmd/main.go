package main

import (
	"github.com/joho/godotenv"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
	"github.com/paulrouge/icon-validator-monitor/internal/icon"
	"github.com/paulrouge/icon-validator-monitor/internal/model"
	"github.com/paulrouge/icon-validator-monitor/internal/sender/tg"
)

// todo:
// - make the validatormap update every N hours
// - get the omm delegation info and implement that info


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
	service := NewMainService(db, tgBot, []model.Sender{}, client)
	service.registerSender(tgBot)

	select{}
}

func NewMainService(db *db.DB, tgBot *tg.TelegramBot, senders []model.Sender, c *icon.Icon) *MainService {
	return &MainService{
		db: db,
		TgBot: tgBot,
		Senders: senders,
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