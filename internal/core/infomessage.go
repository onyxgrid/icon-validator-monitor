package core

import (
	"fmt"
	"strconv"
	"time"

	"github.com/paulrouge/icon-validator-monitor/internal/db"
	"github.com/paulrouge/icon-validator-monitor/internal/icon"
	"github.com/paulrouge/icon-validator-monitor/internal/util"
)

func (e *Engine) SendWeeklyReport() {
	for {
		if time.Now().Weekday() == time.Saturday {
			// get time now, calculate the time until the next 10:00, and sleep for that time
			now := time.Now()
			year, month, day := now.Date()
			next := time.Date(year, month, day, 10, 0, 0, 0, now.Location())
			if now.After(next) {
				next = next.AddDate(0, 0, 1)
			}
			time.Sleep(next.Sub(now))

			uids, err := db.DBInstance.GetAllUserIDs()
			if err != nil {
				e.logger.Error("failed to get all user ids", err)
				return
			}

			for _, uid := range uids {
				uids := strconv.FormatInt(uid, 10)
				wallets := db.DBInstance.GetUserWallets(uids)
				msg := "Weekly Report\n\n"
				// of each wallet, check if it is delegated to the jailed validator
				for _, w := range wallets {
					f := fmt.Sprintf("%s...%s\n", w[:6], w[len(w)-6:])
					msg += fmt.Sprintf("*WALLET* - [%s](https://icontracker.xyz/address/%s)\n", f, w)

					// get the delegation info
					delegation, err := e.Icon.GetDelegation(w)
					if err != nil {
						e.logger.Error("failed to get delegation info: " + err.Error())
						return
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
							e.logger.Error("d: failed to estimate reward: " + err.Error())
							continue
						}
						msg += fmt.Sprintf("Est. daily reward: `$%s`\n\n", util.FormatIconNumber(edr))

					}

					// get the omm votes
					omm := e.Icon.GetOmmVotes(w)

					if len(omm) > 0 {
						msg += "`OMM votes:`\n"
					}

					// for each omm vote, add the address and value to the message
					for _, o := range omm {
						fl := util.FormatIconNumber(o.VotesInIcx)

						msg += fmt.Sprintf("Validator: [%s](https://icontracker.xyz/address/%s)\nOMM votes: `%s ICX`\n", o.Name, o.Address, fl)

						msg += fmt.Sprintf("Commision Rate: `%v%%`\n", e.Validators[o.Address].CommissionRate)

						edr, err := icon.EstimateReward(e.Validators[o.Address], o.VotesInIcx)
						if err != nil {
							e.logger.Error("o: failed to estimate reward: " + err.Error())
							continue
						}
						msg += fmt.Sprintf("Est. daily reward: `$%s`\n\n", util.FormatIconNumber(edr))
					}

					// get the bonds
					bond, err := e.Icon.GetBonds(w)
					if err != nil {
						e.logger.Error("failed to get bond info: " + err.Error())
						return
					}

					if len(bond.Bonds) > 0 {
						msg += "`Bonds:`\n"
					}

					// for each bond, add the address and value to the message
					for _, b := range bond.Bonds {
						fl := util.FormatIconNumber(b.Value)
						msg += fmt.Sprintf("Validator: [%s](https://icontracker.xyz/address/%s)\nBond: `%s` ICX\n", b.Name, b.Address, fl)

						msg += fmt.Sprintf("Commision Rate: `%v%%`\n", e.Validators[b.Address].CommissionRate)

						edr, err := icon.EstimateReward(e.Validators[b.Address], b.Value)
						if err != nil {
							e.logger.Error("b: failed to estimate reward: " + err.Error())
							continue
						}
						msg += fmt.Sprintf("Est. daily reward: `$%s`\n\n", util.FormatIconNumber(edr))
					}
				}

				// send message to all senders
				for _, s := range e.Senders {
					err := s.SendMessage(s.GetReceiver(uids), msg)
					if err != nil {
						e.logger.Error("failed to send message: " + err.Error())
					}
				}
			}

		}
		time.Sleep(24 * time.Hour)
	}
}
