package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/joho/godotenv"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
)

// quick and dirty program to do some custom stuff, check the code for whats happening
//
// in short, we instantiate a tg bot and a connection to the db, 
// to send a message to all users in the db.
//
// but you can adjust this to do whatever you want without interfering with the main bot.

func main() {
	godotenv.Load()

	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		panic("TELEGRAM_TOKEN is not set")
	}

	// Create a Bot client
	b, err := gotgbot.NewBot(token, nil)
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	err = db.NewDB()
	if err != nil {
		panic(err)
	}
	defer db.DBInstance.Close()

	err = db.DBInstance.Init()
	if err != nil {
		panic(err)
	}

	uids, err := db.DBInstance.GetAllUserIDs()
	if err != nil {
		panic(err)
	}

	for _, uid := range uids {
		uids := strconv.Itoa(int(uid))
		u, err := db.DBInstance.GetUser(uids)
		if err != nil {
			fmt.Println(err)
		}

		if u.Inactive {
			continue
		}

		msg := "📢We have updated The Icon Validator Monitor.\n\nYou can now receive CPS alerts. You can toggle this option on/off with the /cps option in the menu.\n\nWhen the CPS alerts are set to active, you will receive alerts when CPS validators are in danger of missing the voting deadline.\n\nIf you have questions or suggestion please open an issue at:\nhttps://github.com/onyxgrid/icon-validator-monitor/issues"

		_, err = b.SendMessage(int64(u.ID), msg, nil)
		if err != nil {
			if strings.Contains(err.Error(), "bot was blocked") {
				fmt.Println("User inactive")
				db.DBInstance.SetUserInactive(uids, true)
			} else {
				fmt.Println(err)
			}
		}
	}

}
