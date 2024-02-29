package core

import (
	"fmt"
	"strconv"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
	"github.com/paulrouge/icon-validator-monitor/internal/icon"
)

// registerWallet registers a wallet
func (t *Engine) registerWallet(b *gotgbot.Bot, ctx *ext.Context) error {
	// Reply to the user
	msg, err := ctx.EffectiveMessage.Reply(b, "Give me the address you want to register, please.", &gotgbot.SendMessageOpts{
		ParseMode: "html",
		ReplyMarkup: &gotgbot.ForceReply{
			ForceReply: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send reply message: %w", err)
	}

	// Save the message ID
	t.registerWalletMsgId = &msg.MessageId

	return nil
}

// handleReply handles the reply from the user
func (t *Engine) handleRegisterReply(ctx *ext.Context) error {
	msg := ctx.EffectiveMessage.Text
	chatID := ctx.EffectiveMessage.Chat.Id
	
	// check if the message is a valid ICON wallet address
	if !icon.IsValidIconAddress(msg) {
		err := t.SendMessage(strconv.FormatInt(chatID, 10), msg + " is not a valid ICON wallet address")
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		
		t.registerWalletMsgId = nil

		return nil
	} else {
		// users current registered wallets
		wallets := db.DBInstance.GetUserWallets(strconv.FormatInt(chatID, 10))

		// check if the wallet is already registered
		for _, wallet := range wallets {
			if wallet == msg {
				err := t.SendMessage(strconv.FormatInt(chatID, 10), msg + " is already registered.")
				if err != nil {
					t.registerWalletMsgId = nil
					return fmt.Errorf("failed to send message: %w", err)
				}

				t.registerWalletMsgId = nil

				return nil
			}
		}

		// add the wallet to the database
		err := db.DBInstance.AddUserWallet(strconv.FormatInt(chatID, 10), msg)
		if err != nil {
			t.registerWalletMsgId = nil
			return fmt.Errorf("failed to add wallet to the database: %w", err)
		}
		
		// Send the message to the chat
		err = t.SendMessage(strconv.FormatInt(chatID, 10), msg + " has been registered.")
		if err != nil {
			t.registerWalletMsgId = nil
			return fmt.Errorf("failed to send message: %w", err)
		}

		// Reset the registerWalletMsgId
		t.registerWalletMsgId = nil

		return nil
	}
}

func (t *Engine) removeWallet(b *gotgbot.Bot, ctx *ext.Context) error {
	// reply to the user
	msg, err := ctx.EffectiveMessage.Reply(b, "Give me the address you want to remove, please.", &gotgbot.SendMessageOpts{
		ParseMode: "html",
		ReplyMarkup: &gotgbot.ForceReply{
			ForceReply: true,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to send reply message: %w", err)
	}

	// Save the message ID
	t.removeWalletMsgId = &msg.MessageId

	return nil
}

func (t *Engine) handleRemoveReply(ctx *ext.Context) error {
	msg := ctx.EffectiveMessage.Text
	chatID := ctx.EffectiveMessage.Chat.Id
	// users current registered wallets
	wallets := db.DBInstance.GetUserWallets(strconv.FormatInt(chatID, 10))

	// check if the wallet is already registered
	for _, wallet := range wallets {
		if wallet == msg {
			// remove the wallet from the database
			err := db.DBInstance.RemoveUserWallet(strconv.FormatInt(chatID, 10), msg)
			if err != nil {
				t.removeWalletMsgId = nil
				return fmt.Errorf("failed to remove wallet from the database: %w", err)
			}
			
			// Send the message to the chat
			err = t.SendMessage(strconv.FormatInt(chatID, 10), msg + " has been removed.")
			if err != nil {
				t.removeWalletMsgId = nil
				return fmt.Errorf("failed to send message: %w", err)
			}

			// Reset the registerWalletMsgId
			t.removeWalletMsgId = nil

			return nil
		}
	}

	// Send the message to the chat
	err := t.SendMessage(strconv.FormatInt(chatID, 10), msg + " is unregistered.")
	if err != nil {
		t.removeWalletMsgId = nil
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Reset the removeWalletMsgId
	t.removeWalletMsgId = nil

	return nil
}