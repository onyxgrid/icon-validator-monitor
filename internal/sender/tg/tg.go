package tg

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/paulrouge/icon-validator-monitor/internal/icon"
	"github.com/paulrouge/icon-validator-monitor/internal/model"
	"github.com/paulrouge/icon-validator-monitor/internal/util"

	// "github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/paulrouge/icon-validator-monitor/internal/db"
)

// to do: maybe have some sort of channel that holds ids of messages that are been sent, and remove them after a certain time

// Bot is the Telegram bot, but i think we need to change it to acutally something like the mainservice or something
// because 
type TelegramBot struct {
	bot *gotgbot.Bot
	registerWalletMsgId *int64
	removeWalletMsgId *int64
	DB *db.DB
	Icon *icon.Icon
	Validators map[string]model.ValidatorInfo
}

// NewBot creates a new Bot
func NewBot(d *db.DB, i *icon.Icon) (*TelegramBot, error) {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN is not set")
	}

	var b *gotgbot.Bot

	// Create a Bot client
	b, err := gotgbot.NewBot(token, nil)

	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	validators, err := i.GetAllValidators()
	if err != nil {
		return nil, fmt.Errorf("failed to get validators: %w", err)
	}

	validatorsMap := make(map[string]model.ValidatorInfo)
	for _, v := range validators {
		validatorsMap[v.Address] = v
	}


	return &TelegramBot{bot: b, DB: d, Icon: i, Validators: validatorsMap}, nil
}

// Init initializes the bot
func (t *TelegramBot) Init() {

	// Create updater and dispatcher.
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		// If an error is returned by a handler, log it and continue going.
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	
	// /start command to introduce the bot
	dispatcher.AddHandler(handlers.NewCommand("start", t.start))
	// /register command to register a wallet
	dispatcher.AddHandler(handlers.NewCommand("register", t.registerWallet))
	// /wallets command to show the wallets of a user
	dispatcher.AddHandler(handlers.NewCommand("mywallets", t.showWallets))
	// /remove command to remove a wallet
	dispatcher.AddHandler(handlers.NewCommand("remove", t.removeWallet))

	updater := ext.NewUpdater(dispatcher, nil)

	// Add echo handler to reply to all text messages.
	dispatcher.AddHandler(handlers.NewMessage(message.Text, t.Listen))
	
	// Start receiving updates.
	err := updater.StartPolling(t.bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		panic("failed to start polling: " + err.Error())
	}
	log.Printf("%s has been started...\n", t.bot.User.Username)

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()

	// return nil
}

// Listen listens to messages
func (t *TelegramBot)Listen(b *gotgbot.Bot, ctx *ext.Context) error {
	// check if the message is a reply
	if ctx.EffectiveMessage.ReplyToMessage != nil {
		// check if the message is a reply to the registerWallet message
		if t.registerWalletMsgId != nil && ctx.EffectiveMessage.ReplyToMessage.MessageId == *t.registerWalletMsgId {
			return t.handleRegisterReply(b, ctx)
		}
		if t.removeWalletMsgId != nil && ctx.EffectiveMessage.ReplyToMessage.MessageId == *t.removeWalletMsgId {
			return t.handleRemoveReply(b, ctx)
		}
	}
	
	return nil
}

// start introduces the bot.
func (t *TelegramBot) start(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Welcome, I am %s.\n\nType /help to get an overview of what I can help you with.", b.User.Username), &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}

	// add the user to the database
	err = t.DB.AddUser(strconv.FormatInt(ctx.EffectiveMessage.Chat.Id, 10))
	if err != nil {
		return fmt.Errorf("failed to add user to the database: %w", err)
	}

	return nil
}

// SendMessage sends a message to a chat
func (t *TelegramBot) SendMessage(chatID string, message string) error {
	// str to int64
	i, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return err
	}

	opts := &gotgbot.SendMessageOpts{
		ParseMode: "Markdown",
	}

	_, err = t.bot.SendMessage(i, message, opts)
	if err != nil {
		return err
	}
	return nil
}

// registerWallet registers a wallet
func (t *TelegramBot) registerWallet(b *gotgbot.Bot, ctx *ext.Context) error {
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
func (t *TelegramBot) handleRegisterReply(b *gotgbot.Bot, ctx *ext.Context) error {
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
		wallets := t.DB.GetUserWallets(strconv.FormatInt(chatID, 10))

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
		err := t.DB.AddUserWallet(strconv.FormatInt(chatID, 10), msg)
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


// showWallets shows the wallets of a user, and the delegation info
func (t *TelegramBot) showWallets(b *gotgbot.Bot, ctx *ext.Context) error {
	chatID := ctx.EffectiveMessage.Chat.Id
	wallets := t.DB.GetUserWallets(strconv.FormatInt(chatID, 10))
	if wallets == nil {
		err := t.SendMessage(strconv.FormatInt(chatID, 10), "You have no registered wallets.")
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	msg := ""
	
	for _, wallet := range wallets {
		// format address to hx012...h921
		f := fmt.Sprintf("%s...%s\n", wallet[:6], wallet[len(wallet)-6:])
		msg += fmt.Sprintf("*WALLET* - [%s](https://icontracker.xyz/address/%s)\n", f, wallet)

		// get the delegation info
		delegation, err := t.Icon.GetDelegation(wallet)
		if err != nil {
			return fmt.Errorf("failed to get delegation info: %w", err)
		}

		if len(delegation.Delegations) > 0 {
			msg += "`Regular votes`:\n"
		}

		// for each delegation, add the address and value to the message
		for _, d := range delegation.Delegations {
			fl := util.FormatIconNumber(d.Value)
			msg += fmt.Sprintf("Validator: [%s](https://icontracker.xyz/address/%s)\nvotes: `%s` ICX\n", d.Name, d.Address, fl)

			msg += fmt.Sprintf("Commision Rate: `%v%%`\n", t.Validators[d.Address].CommissionRate)
		
			edr, err := icon.EstimateReward(t.Validators[d.Address], d.Value)
			if err != nil {
				continue
			}
			msg += fmt.Sprintf("Est. daily reward: `$%s`\n\n", util.FormatIconNumber(edr))
			
		}
		
		omm := t.Icon.GetOmmVotes(wallet)

		if len(omm) > 0 {
			msg += "`OMM votes:`\n"
		}

		for _, o := range omm {
			fl := util.FormatIconNumber(o.VotesInIcx)

			msg += fmt.Sprintf("Validator: [%s](https://icontracker.xyz/address/%s)\nOMM votes: `%s ICX`\n", o.Name, o.Address, fl)
			
			msg += fmt.Sprintf("Commision Rate: `%v%%`\n", t.Validators[o.Address].CommissionRate)

			edr, err := icon.EstimateReward(t.Validators[o.Address], o.VotesInIcx)
			if err != nil {
				continue
			}
			msg += fmt.Sprintf("Est. daily reward: `$%s`\n\n", util.FormatIconNumber(edr))
		}
		
		// get the bond info
		bond, err := t.Icon.GetBonds(wallet)
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
			
			edr, err := icon.EstimateReward(t.Validators[b.Address], b.Value)
			if err != nil {
				continue
			}
			msg += fmt.Sprintf("Est. daily reward: `$%s`\n\n", util.FormatIconNumber(edr))
		}
	}

	// Send the message to the chat
	err := t.SendMessage(strconv.FormatInt(chatID, 10), msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (t *TelegramBot) removeWallet(b *gotgbot.Bot, ctx *ext.Context) error {
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

func (t *TelegramBot) handleRemoveReply(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage.Text
	chatID := ctx.EffectiveMessage.Chat.Id
	// users current registered wallets
	wallets := t.DB.GetUserWallets(strconv.FormatInt(chatID, 10))

	// check if the wallet is already registered
	for _, wallet := range wallets {
		if wallet == msg {
			// remove the wallet from the database
			err := t.DB.RemoveUserWallet(strconv.FormatInt(chatID, 10), msg)
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