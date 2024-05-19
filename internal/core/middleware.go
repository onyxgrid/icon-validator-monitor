package core

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
)

// authHandler makes sure the user exists in the database
func (e *Engine) authHandler(h handlers.Response) handlers.Response {
	return func(b *gotgbot.Bot, ctx *ext.Context) error {
		uid := strconv.Itoa(int(ctx.EffectiveMessage.Chat.Id))
		
		// add user to db if not exists
		err := db.DBInstance.AddUser(uid)
		if err != nil {
			fmt.Println("Error adding user to db: ", err)
		}

		// set to users inactive state to false (so if a user reactivates the bot, they will get notifications)
		db.DBInstance.SetUserInactive(uid, false)

		middelwareLogger(ctx.EffectiveMessage.Text + " - " + uid)
		return h(b, ctx)
	}
}

func middelwareLogger(msg string) {
	logFile, err := os.OpenFile("data/middleware.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	slog := slog.New(slog.NewTextHandler(logFile, nil))
	slog.Info(msg)
}
