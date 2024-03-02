package core

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

	"github.com/paulrouge/icon-validator-monitor/internal/db"
)


type Engine struct {
	bot *gotgbot.Bot
	registerWalletMsgId *int64
	removeWalletMsgId *int64
	setEmailAddrMsgId *int64
	
	Icon *icon.Icon
	Validators map[string]model.ValidatorInfo

	Senders []model.Sender
}

func NewEngine(d *db.DB, i *icon.Icon) (*Engine, error) {
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

	return &Engine{bot: b, Icon: i, Validators: validatorsMap}, nil
}

func (t *Engine) RegisterSender(s model.Sender) {
	t.Senders = append(t.Senders, s)
}

func (t *Engine) GetReceiver(uid string) string {
	// uid and receiver are the same for core sender
	return uid
}

func (t *Engine) Init() {
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
	// /setemail command to set the email address
	dispatcher.AddHandler(handlers.NewCommand("setemail", t.setEmailAddr))

	// Handle all text messages.
	dispatcher.AddHandler(handlers.NewMessage(message.Text, t.Listen))
	
	updater := ext.NewUpdater(dispatcher, nil)
	
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

	updater.Idle()
}

// Listen listens to messages
func (t *Engine)Listen(b *gotgbot.Bot, ctx *ext.Context) error {
	// check if the message is a reply
	if ctx.EffectiveMessage.ReplyToMessage != nil {
		// check if the message is a reply to the registerWallet message
		if t.registerWalletMsgId != nil && ctx.EffectiveMessage.ReplyToMessage.MessageId == *t.registerWalletMsgId {
			return t.handleRegisterReply(ctx)
		}
		if t.removeWalletMsgId != nil && ctx.EffectiveMessage.ReplyToMessage.MessageId == *t.removeWalletMsgId {
			return t.handleRemoveReply(ctx)
		}
		if t.setEmailAddrMsgId != nil && ctx.EffectiveMessage.ReplyToMessage.MessageId == *t.setEmailAddrMsgId {
			return t.handleSetEmailAddrReply(ctx)
		}
	}
	
	return nil
}

// start introduces the bot.
func (t *Engine) start(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Welcome, I am %s.\n\nType /help to get an overview of what I can help you with.", b.User.Username), &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}

	// add the user to the database
	err = db.DBInstance.AddUser(strconv.FormatInt(ctx.EffectiveMessage.Chat.Id, 10))
	if err != nil {
		return fmt.Errorf("failed to add user to the database: %w", err)
	}

	return nil
}

// SendMessage sends a message to a chat
func (t *Engine) SendMessage(chatID string, message string) error {
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

// SendAlert sends an alert to a user
func (t *Engine) SendAlert(chatID string, v string, w string) error {
	i, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return err
	}

	opts := &gotgbot.SendMessageOpts{
		ParseMode: "Markdown",
	}

	msg := fmt.Sprintf("Validator jailed: *%s*\n%s is not earning rewards for the ICX delegated to this validator!", v, w)

	_, err = t.bot.SendMessage(i, msg, opts)
	if err != nil {
		return err
	}
	return nil
}

// UpdateValidators updates the validatormap every hour
func (t *Engine) UpdateValidators() {
	for {
		validators, err := t.Icon.GetAllValidators()
		if err != nil {
			log.Println("failed to get validators: " + err.Error())
			continue
		}

		validatorsMap := make(map[string]model.ValidatorInfo)
		for _, v := range validators {
			validatorsMap[v.Address] = v
		}

		t.Validators = validatorsMap
		t.checkJail()

		time.Sleep(time.Hour)
	}
}

