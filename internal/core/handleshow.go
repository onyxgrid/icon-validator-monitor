package core

import (
	"fmt"
	"strconv"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
	"github.com/paulrouge/icon-validator-monitor/internal/icon"
	"github.com/paulrouge/icon-validator-monitor/internal/util"
)

// showWallets shows the wallets of a user, and the delegation info
func (e *Engine) showWallets(b *gotgbot.Bot, ctx *ext.Context) error {
	chatID := ctx.EffectiveMessage.Chat.Id

	u, err := db.DBInstance.GetUser(strconv.FormatInt(chatID, 10))
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if len(u.Wallets) == 0 {
		err := e.SendMessage(strconv.FormatInt(chatID, 10), "You have no registered wallets.")
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	msg := ""

	for _, wallet := range u.Wallets {
		if wallet == "" {
			continue
		}
		
		// format address to hx012...h921
		f := fmt.Sprintf("%s...%s\n", wallet[:6], wallet[len(wallet)-6:])
		msg += fmt.Sprintf("*WALLET* - [%s](https://icontracker.xyz/address/%s)\n", f, wallet)

		// get the delegation info
		delegation, err := e.Icon.GetDelegation(wallet)
		if err != nil {
			return fmt.Errorf("failed to get delegation info for address %v: %w", wallet, err)
		}

		if len(delegation.Delegations) > 0 {
			msg += "`Regular votes`:\n"
		}

		// for each delegation, add the address and value to the message
		for _, d := range delegation.Delegations {
			fl := util.FormatIconNumber(d.Value)
			msg += fmt.Sprintf("Validator: [%s](https://icontracker.xyz/address/%s)\nvotes: `%s` ICX\n", d.Name, d.Address, fl)

			msg += fmt.Sprintf("Commision Rate: `%v%%`\n", e.Validators[d.Address].CommissionRate)

			edr, err := icon.EstimateReward(e.Validators[d.Address], d.Value)
			if err != nil {
				continue
			}
			msg += fmt.Sprintf("Est. daily reward: `$%s`\n\n", util.FormatIconNumber(edr))

		}

		// get the omm votes
		omm := e.Icon.GetOmmVotes(wallet)

		if len(omm) > 0 {
			msg += "`OMM votes:`\n"
		}

		for _, o := range omm {
			fl := util.FormatIconNumber(o.VotesInIcx)

			msg += fmt.Sprintf("Validator: [%s](https://icontracker.xyz/address/%s)\nOMM votes: `%s ICX`\n", o.Name, o.Address, fl)

			/*
				This could be custimonized to show the custom rewards for each validator
				and extend msg with the custom rewards
			*/
		}

		// get the bond info
		bond, err := e.Icon.GetBonds(wallet)
		if err != nil {
			return fmt.Errorf("failed to get bond info: %w", err)
		}

		if len(bond.Bonds) > 0 {
			msg += "`Bonds:`\n"
		}

		// for each bond, add the address and value to the message
		for _, b := range bond.Bonds {
			fl := util.FormatIconNumber(b.Value)
			msg += fmt.Sprintf("Validator: [%s](https://icontracker.xyz/address/%s)\nBonded: `%s ICX`\n", b.Name, b.Address, fl)

			edr, err := icon.EstimateReward(e.Validators[b.Address], b.Value)
			if err != nil {
				continue
			}
			msg += fmt.Sprintf("Est. daily reward: `$%s`\n\n", util.FormatIconNumber(edr))
		}
	}

	// Send the message to the chat
	err = e.SendMessage(strconv.FormatInt(chatID, 10), msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
