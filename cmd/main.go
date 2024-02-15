package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
	"github.com/paulrouge/icon-validator-monitor/internal/icon"
	"github.com/paulrouge/icon-validator-monitor/internal/model"
	"github.com/paulrouge/icon-validator-monitor/internal/sender/tg"
)


type MainService struct {
	db *db.DB
	TgBot *tg.TelegramBot
	Senders []model.Sender
	Icon *icon.Icon
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

	// Create a new Telegram bot
	tgBot, err := tg.NewBot(db); if err != nil {
		panic(err)
	}

	go tgBot.Init();

	// Create a new Icon client
	client, err := icon.NewIcon(); if err != nil {
		panic(err)
	}

	adr := "hxf0fa3e948b05eca500400a88c13808d0797e4850"
	r, err := client.GetDelegation(adr)
	if err != nil {
		panic(err)
	}

	for _, d := range r.Delegations {
		fmt.Println(d.Address, d.Value, d.Name)
	}
	
	// Create a new MainService
	service := NewMainService(db, tgBot, []model.Sender{}, client)
	_ = service

	select{}
	
}