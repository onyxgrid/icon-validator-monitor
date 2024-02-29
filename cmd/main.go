package main

import (
	"github.com/joho/godotenv"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
	"github.com/paulrouge/icon-validator-monitor/internal/icon"
	"github.com/paulrouge/icon-validator-monitor/internal/sender/mail"
	"github.com/paulrouge/icon-validator-monitor/internal/core"
)

// todo:
// - Add correct logging troughout the code
// - Create a function that sends a weekly report to all users

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

	engine, err := core.NewEngine(db.DBInstance, client); if err != nil {
		// this should be added to the log. (failed on parsing res into validotrinfo before...)	
		panic(err)
	}

	go engine.Init();
	
	// Register the senders that will send the notifications
	engine.RegisterSender(engine)

	// Create a gmail sender
	gmailSender, err := mail.NewMail(); if err != nil {
		panic(err)
	}

	// Register the gmail sender
	engine.RegisterSender(gmailSender)

	// update the validators every hour
	go engine.UpdateValidators()
	
	select{}
}
