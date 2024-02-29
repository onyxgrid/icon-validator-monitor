package core

import (
	"fmt"
	"log"
	"strconv"

	"github.com/paulrouge/icon-validator-monitor/internal/db"
)

func (t *Engine) checkJail() {
	// for testing set hx2e7db537ca3ff73336bee1bab4cf733a94ae769b to jail_flag 0x1
	x := t.Validators["hx2e7db537ca3ff73336bee1bab4cf733a94ae769b"]
	x.JailFlags = "0x1"

	// set x to the validators map
	t.Validators["hx2e7db537ca3ff73336bee1bab4cf733a94ae769b"] = x

	// check for jail_flag
	for a, v := range t.Validators {
		if v.JailFlags != "0x0" {
			UIDs, err := db.DBInstance.GetAllUserIDs()
			if err != nil {
				log.Println("failed to get all users: " + err.Error())
				return
			}

			for _, uid := range UIDs {
				uids := strconv.FormatInt(uid, 10)
				wallets := db.DBInstance.GetUserWallets(uids)

				// of each wallet, check if it is delegated to the jailed validator
				for _, w := range wallets {
					// check regular votes
					delegation, err := t.Icon.GetDelegation(w)
					if err != nil {
						log.Println("failed to get delegation info: " + err.Error())
						return
					}
					
					for _, d := range delegation.Delegations {
						if d.Address == a {
							err := t.SendMessage(uids, "Validator jailed: " + v.Name)
							if err != nil {
								log.Println("failed to send message: " + err.Error())
							}
						}
					}

					// check omm votes
					omm := t.Icon.GetOmmVotes(w)
					for _, o := range omm {
						if o.Address == a {
							for _, s := range t.Senders {
								msg := fmt.Sprintf("Validator jailed: %s\nYou are not earning any rewards as long as the validator is jailed.", v.Name)

								r := s.GetReceiver(uids)
								fmt.Println("sending message to: " + r)
								if r == "" {
									continue
								}

								
								err := s.SendMessage(r, msg)
								if err != nil {
									log.Println("failed to send message: " + err.Error())
								}
							}
						}
					}

					// check bonds
					bond, err := t.Icon.GetBonds(w)
					if err != nil {
						log.Println("failed to get bond info: " + err.Error())
						return
					}

					for _, b := range bond.Bonds {
						if b.Address == a {
							err := t.SendMessage(uids, "Validator jailed: " + v.Name)
							if err != nil {
								log.Println("failed to send message: " + err.Error())
							}
						}
					}
				}
			}
		}
	}

}