package core

import (
	"fmt"
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
		err := db.DBInstance.AddUser(uid)
		if err != nil {
			fmt.Println("Error adding user to db: ", err)
		}
		return h(b, ctx)
	}
}
