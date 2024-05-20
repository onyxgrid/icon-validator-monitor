package core

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
)

func (e *Engine) toggleCPSAlert(b *gotgbot.Bot, ctx *ext.Context) error {
	uid := int(ctx.EffectiveMessage.Chat.Id)
	uids := strconv.Itoa(uid)

	if uids == "" {
		fmt.Println("failed to get user ID")
		return fmt.Errorf("failed to get user ID")
	}

	u, err := db.DBInstance.GetUser(uids)
	if err != nil {
		fmt.Println("failed to get user", err)
		return fmt.Errorf("failed to get user: %w", err)
	}

	var reply string

	// Check if the user has already registered an address
	if slices.Contains(u.Alerts, "CPS") {
		// Remove the alert
		db.DBInstance.RemoveAlert(uids, "CPS")
		reply = "CPS alert removed."
	} else {
		err := db.DBInstance.AddAlert(uids, "CPS")
		if err != nil {
			fmt.Println("failed to add alert", err)
			return fmt.Errorf("failed to add alert: %w", err)
		}
		reply = "CPS alert added.\n\nYou will get a notification when a CPS register validator is about to miss a CPS vote."
	}

	// Reply to the user
	_, err = b.SendMessage(ctx.EffectiveMessage.Chat.Id, reply, nil)
	if err != nil {
		return fmt.Errorf("failed to send cps message: %w", err)
	}

	return nil
}
