package core

import (
	"log"
	"strconv"

	"github.com/paulrouge/icon-validator-monitor/internal/db"
)

func (t *Engine) checkJail() {
	// for testing set hx2e7db537ca3ff73336bee1bab4cf733a94ae769b to jail_flag 0x1
	// x := t.Validators["hx2e7db537ca3ff73336bee1bab4cf733a94ae769b"]
	// x.JailFlags = "0x1"
	// t.Validators["hx2e7db537ca3ff73336bee1bab4cf733a94ae769b"] = x

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

				u, err := db.DBInstance.GetUser(uids)
				if err != nil {
					log.Println("failed to get user: " + err.Error())
					return
				}

				// of each wallet, check if it is delegated to the jailed validator
				for _, w := range u.Wallets {
					if w == "" {
						continue
					}
					// check regular votes
					delegation, err := t.Icon.GetDelegation(w)
					if err != nil {
						log.Println("reg votes - failed to get delegation info for wallet: " + w + err.Error())
						return
					}

					for _, d := range delegation.Delegations {
						if d.Address == a {
							err := t.SendAlerts(uids, v.Name, w)
							if err != nil {
								log.Println("failed to send alert: " + err.Error())
							}
						}
					}

					// check omm votes
					omm := t.Icon.GetOmmVotes(w)
					for _, o := range omm {
						if o.Address == a {
							err := t.SendAlerts(uids, v.Name, w)
							if err != nil {
								log.Println("failed to send alert: " + err.Error())
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
							err := t.SendAlerts(uids, v.Name, w)
							if err != nil {
								log.Println("failed to send alert: " + err.Error())
							}
						}
					}
				}
			}
		}
	}
}

// send alerts to all senders
func (t *Engine) SendAlerts(chatID string, val string, w string) error {
	for _, s := range t.Senders {
		err := s.SendAlert(s.GetReceiver(chatID), val, w)
		if err != nil {
			t.Logger.Error("failed to send alert", err, "chatID: ", chatID, "validator: ", val, "wallet: ", w)
			return err
		}
	}
	return nil
}
