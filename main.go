package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"sync"
)

const (
	admin        = "yaspe"
	dataFileName = "data.db"
	betTimeFrom  = 15
	betTimeTo    = 7
)

var (
	current          = 0
	shouldShutdown   = false
	bets             = make(map[string]int)
	awaitingBets     = make(map[int64]bool)
	chats            = make(map[string]int64)
	winners          = make(map[string]int)
	dataLock         sync.Mutex
	forceBetsAllowed = true
)

var version = "unknown" // -ldflags "-X 'main.version=`git rev-list --all --count`'"
var bot *tgbotapi.BotAPI

func main() {
	Load()
	go dumpLoop()

	token := os.Args[1]

	var er error
	bot, er = tgbotapi.NewBotAPI(token)
	if er != nil {
		log.Panic(er)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, er := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.ForwardFrom != nil && update.Message.ForwardFrom.UserName == admin {
			for _, chat := range chats {
				fwd := tgbotapi.NewForward(chat, chats[admin], update.Message.MessageID)
				_, er = bot.Send(fwd)
				if er != nil {
					log.Printf("Could not send message: %s", er)
				}
			}
			continue
		}

		var msg tgbotapi.MessageConfig

		result := handleMessage(update.Message)
		if shouldShutdown {
			dataLock.Lock()
			return
		}
		if result.Error != nil {
			log.Printf("Could not process message: %s %s", update.Message.Text, result.Error)
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, result.Error.Error())
			//msg.ReplyToMessageID = update.Message.MessageID
		} else {
			if result.BroadCast {
				for username, chat := range chats {
					if result.RemindMode {
						if val, ok := bets[username]; ok && val > 0 {
							continue
						}
					}
					log.Printf("Broadcasting message to: %d", chat)
					msg = tgbotapi.NewMessage(chat, result.Reply)
					_, er := bot.Send(msg)
					if er != nil {
						log.Printf("Could not send message: %s", er)
					}
				}
				continue
			} else {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, result.Reply)
			}
		}

		markup := tgbotapi.ReplyKeyboardMarkup{
			Keyboard: [][]tgbotapi.KeyboardButton{
				{
					{Text: "/bet"},
					{Text: "/mybet"},
					{Text: "/winners"},
					{Text: "/help"},
				},
			},
			ResizeKeyboard: true,
		}
		msg.ReplyMarkup = markup

		_, er = bot.Send(msg)
		if er != nil {
			log.Printf("Could not send message: %s", er)
		}
	}
}
