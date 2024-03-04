package core

import (
	"fmt"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
)

func (e *Engine) handleTestSenders(b *gotgbot.Bot, ctx *ext.Context) error {
	uid := fmt.Sprintf("%d", ctx.EffectiveMessage.Chat.Id)

	msg := "A test-alert will be send.\n\n"
	email := db.DBInstance.GetUserEmail(uid)
	if email == "" {
		msg += "You haven't set an email address yet.\nThe test-alert will be send in this telegram chat in 10 seconds. Please use /setemail if you also want to receive email alerts."
	} else {
		msg += fmt.Sprintf("A test-alert will be send to your email address: %s and to this telegram chat in 10 seconds", email)
	}

	err := e.SendMessage(uid, msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	time.Sleep(10 * time.Second)

	for _, sender := range e.Senders {
		err = sender.SendMessage(sender.GetReceiver(uid), "This is a test alert.")
		if err != nil {
			return fmt.Errorf("failed to send test alert: %w", err)
		}
	}

	return nil
}
