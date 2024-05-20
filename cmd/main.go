package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/paulrouge/icon-validator-monitor/internal/core"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
	"github.com/paulrouge/icon-validator-monitor/internal/icon"
	"github.com/paulrouge/icon-validator-monitor/internal/sender/mail"
)

/* helper for the botfather commands, you can paste these in when setting menu command in the botfather menu:

register - Register a wallet to track
cps - Receive CPS alerts
remove - Remove a wallet from tracking
mywallets - Get the statistics of all your wallets
setemail - Set the email to receive alerts
testalert - Test the alert system

*/

//todo handle all the send messages errors, set user to inactive if the the tg message is not send, doing it already at some places

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Create a new DB connection
	err = db.NewDB()
	if err != nil {
		panic(err)
	}
	defer db.DBInstance.Close()

	err = db.DBInstance.Init()
	if err != nil {
		panic(err)
	}

	// Create a new Icon client
	client, err := icon.NewIcon()
	if err != nil {
		panic(err)
	}

	// Make logfile
	logFile, err := os.OpenFile("data/log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	engine, err := core.NewEngine(db.DBInstance, client, logFile)
	if err != nil {
		panic(err)
	}

	engine.Logger.Info("Service starting...")

	go engine.Init()

	// Register the senders that will send the notifications
	engine.RegisterSender(engine)

	// Create a gmail sender
	gmailSender, err := mail.NewMail()
	if err != nil {
		panic(err)
	}

	// Register the gmail sender
	engine.RegisterSender(gmailSender)

	// update the validators every hour
	engine.UpdateValidators()

	// send the weekly report every saturday at 10:00
	engine.ScheduleWeekdayTask(6, 10, 0, engine.SendWeeklyReport)

	// the cps service
	engine.RunCPSService()

	select {}
}
