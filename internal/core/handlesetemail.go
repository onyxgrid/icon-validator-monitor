package core

import (
	"fmt"
	"strconv"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
)

func (t *Engine) setEmailAddr(b *gotgbot.Bot, ctx *ext.Context) error {
	// reply to the user
	msg, err := ctx.EffectiveMessage.Reply(b, "Give me the email address you want to set, please.", &gotgbot.SendMessageOpts{
		ParseMode: "html",
		ReplyMarkup: &gotgbot.ForceReply{
			ForceReply: true,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to send reply message: %w", err)
	}

	// Save the message ID
	t.setEmailAddrMsgId = &msg.MessageId

	return nil
}

func (t *Engine) handleSetEmailAddrReply(ctx *ext.Context) error {
	msg := ctx.EffectiveMessage.Text
	chatID := ctx.EffectiveMessage.Chat.Id

	// add the email address to the database
	err := db.DBInstance.SetUserEmail(strconv.FormatInt(chatID, 10), msg)
	if err != nil {
		t.setEmailAddrMsgId = nil
		return fmt.Errorf("failed to add email address to the database: %w", err)
	}

	// Send the message to the chat
	err = t.SendMessage(strconv.FormatInt(chatID, 10), msg + " has been set.")
	if err != nil {
		t.setEmailAddrMsgId = nil
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Reset the setEmailAddrMsgId
	t.setEmailAddrMsgId = nil

	return nil
}