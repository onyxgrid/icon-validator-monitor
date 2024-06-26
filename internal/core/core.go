package core

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
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
	bot                 *gotgbot.Bot
	Logger              *slog.Logger
	registerWalletMsgId *int64
	removeWalletMsgId   *int64
	setEmailAddrMsgId   *int64

	Icon       *icon.Icon
	Validators map[string]model.ValidatorInfo

	Senders []model.Sender
}

func NewEngine(d *db.DB, i *icon.Icon, l *os.File) (*Engine, error) {
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

	Logger := slog.New(slog.NewTextHandler(l, nil))
	validators, err := i.GetAllValidators()
	if err != nil {
		Logger.Error("failed to get validators", err)
		return nil, fmt.Errorf("failed to get validators: %w", err)
	}

	validatorsMap := make(map[string]model.ValidatorInfo)
	for _, v := range validators {
		validatorsMap[v.Address] = v
	}

	return &Engine{bot: b, Icon: i, Validators: validatorsMap, Logger: Logger}, nil
}

func (t *Engine) RegisterSender(s model.Sender) {
	t.Senders = append(t.Senders, s)
}

func (t *Engine) GetReceiver(uid string) string {
	// uid and receiver are the same for core sender
	return uid
}

func (e *Engine) Init() {
	// Create updater and dispatcher.
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		// If an error is returned by a handler, log it and continue going.
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			e.Logger.Error("error creating dispatcher", err)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})

	// /start command to introduce the bot
	dispatcher.AddHandler(handlers.NewCommand("start", e.authHandler(e.start)))
	// /register command to register a wallet
	dispatcher.AddHandler(handlers.NewCommand("register", e.authHandler(e.registerWallet)))
	// /wallets command to show the wallets of a user
	dispatcher.AddHandler(handlers.NewCommand("mywallets", e.authHandler(e.showWallets)))
	// /remove command to remove a wallet
	dispatcher.AddHandler(handlers.NewCommand("remove", e.authHandler(e.removeWallet)))
	// /setemail command to set the email address
	dispatcher.AddHandler(handlers.NewCommand("setemail", e.authHandler(e.setEmailAddr)))
	// /testsenders command to test the senders
	dispatcher.AddHandler(handlers.NewCommand("testalert", e.authHandler(e.handleTestSenders)))
	// /cps command to toggle the CPS alert
	dispatcher.AddHandler(handlers.NewCommand("cps", e.authHandler(e.toggleCPSAlert)))

	// Handle all text messages.
	dispatcher.AddHandler(handlers.NewMessage(message.Text, e.Listen))

	updater := ext.NewUpdater(dispatcher, nil)

	// Start receiving updates.
	err := updater.StartPolling(e.bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		e.Logger.Error("failed to start polling", err)
		panic("failed to start polling: " + err.Error())
	}
	log.Printf("%s has been started...\n", e.bot.User.Username)

	updater.Idle()
}

// Listen listens to messages
func (t *Engine) Listen(b *gotgbot.Bot, ctx *ext.Context) error {
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
func (e *Engine) start(b *gotgbot.Bot, ctx *ext.Context) error {
	// send the introduction message
	msg := "Welcome to the ICON Validator Monitor Bot!\n\n"
	msg += "With this bot you can monitor your ICON wallets.\nRegister wallets that you want to keep track of.\nYou can get an overview of all your registered wallets with /mywallets\n\nYou will also receive a weekly overview every Saturday and his bot will send you an alert if a validator is jailed and not earning rewards for the ICX you delegated to this validator.\n\nSet an email adres if you also want to receive messages via email.\n\n"

	_, err := b.SendMessage(ctx.EffectiveMessage.Chat.Id, msg, nil)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// SendMessage sends a message to a chat
func (e *Engine) SendMessage(chatID string, message string) error {
	u, err := db.DBInstance.GetUser(chatID)
	if err != nil {
		return err
	}

	if u.Inactive {
		return nil
	}
	
	// str to int64
	i, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return err
	}

	opts := &gotgbot.SendMessageOpts{
		ParseMode: "Markdown",
	}

	_, err = e.bot.SendMessage(i, message, opts)
	if err != nil {
		if strings.Contains(err.Error(), "bot was blocked") {
			db.DBInstance.SetUserInactive(chatID, true)
		} else {
			e.Logger.Error("failed to send message", err, "chatID: ", chatID, "message: ", message)
		}
		return err
	}
	return nil
}

// SendAlert sends an alert to a user
func (e *Engine) SendAlert(chatID string, v string, w string) error {
	i, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return err
	}

	opts := &gotgbot.SendMessageOpts{
		ParseMode: "Markdown",
	}

	msg := fmt.Sprintf("Validator jailed: *%s*\n%s is not earning rewards for the ICX delegated to this validator!", v, w)

	//todo handle the error, if the user has blocked the bot the user should be removed from the db.
	_, err = e.bot.SendMessage(i, msg, opts)
	if err != nil {
		if strings.Contains(err.Error(), "bot was blocked") {
			db.DBInstance.SetUserInactive(chatID, true)
		} else {
			e.Logger.Error("failed to send message", err, "chatID: ", chatID, "message: ", msg)
		}
		return err
	}
	return nil
}

// UpdateValidators updates the validatormap every hour
func (e *Engine) UpdateValidators() {
	go func() {
		for {
			validators, err := e.Icon.GetAllValidators()
			if err != nil {
				e.Logger.Error("failed to get validators", err)
				continue
			}

			validatorsMap := make(map[string]model.ValidatorInfo)
			for _, v := range validators {
				validatorsMap[v.Address] = v
			}

			e.Validators = validatorsMap
			e.checkJail()

			time.Sleep(time.Minute)
		}
	}()

}

func (e *Engine) RunCPSService() {	
	go func() {
		for {
			t, err := e.Icon.GetRemainingTimePeriod()
			if err != nil {
				log.Println(err)
				return
			}

			if t < time.Hour*3*24 {
				var priorityMessage string
				var proposalMessage string
				var progressMessage string

				preps, err := e.Icon.GetPreps()
				if err != nil {
					log.Println(err)
					return
				}

				for _, p := range preps {
					priority, err := e.Icon.CheckPriorityVoting(p.Address)
					if err != nil {
						log.Println(err)
						return
					}

					if !priority {
						priorityMessage += fmt.Sprintf("🚨`%s` still have to make the Priority vote!\n\n", p.Name)
					}

					proposal, err := e.Icon.GetRemainingProject(p.Address, "proposal")
					if err != nil {
						log.Println(err)
						return
					}

					if len(proposal) > 0 {
						proposalMessage += fmt.Sprintf("🚨`%s` has %d remaining proposals\n\n", p.Name, len(proposal))
					}

					progress, err := e.Icon.GetRemainingProject(p.Address, "progress_reports")
					if err != nil {
						log.Println(err)
						return
					}

					if len(progress) > 0 {
						progressMessage += fmt.Sprintf("🚨`%s` has %d remaining progress reports\n\n", p.Name, len(progress))
					}
				}

				if priorityMessage == "" && proposalMessage == "" && progressMessage == "" {
					time.Sleep(time.Hour)
					return
				}

				msg := fmt.Sprintf("CPS Service Alert\n\n%s%s%s\nTime Left: %v", priorityMessage, proposalMessage, progressMessage, t)

				e.sendCPSServiceAlert(msg)

			// check every 6 hours if less than 24 hours left
			if t < time.Hour*24 {
				time.Sleep(time.Hour * 6)
			} else {
				time.Sleep(time.Hour * 24)
			}
			}
		}
	}()
}

func (e *Engine) sendCPSServiceAlert(m string) {
	users, err := db.DBInstance.GetUsersPerAlert("CPS")
	if err != nil {
		log.Println(err)
		return
	}

	for _, u := range users {
		us := fmt.Sprintf("%v", u)
		// send message to all senders
		for _, s := range e.Senders {
			err := s.SendMessage(s.GetReceiver(us), m)
			if err != nil {
				e.Logger.Error("failed to send message: " + err.Error())
			}
		}
	}
}
