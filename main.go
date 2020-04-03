package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"sync"
)

const (
	admin = "yaspe"
 	dataFileName = "data.db"
 	betTimeFrom = 15
 	betTimeTo = 7
)

var (
	current          = 0
	shouldShutdown   = false
	bets             = make(map[string]int)
	chats            = make(map[string]int64)
	winners          = make(map[string]int)
	dataLock         sync.Mutex
	forceBetsAllowed = true
)

var bot *tgbotapi.BotAPI

func main() {
	Load()
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

		_, er = bot.Send(msg)
		if er != nil {
			log.Printf("Could not send message: %s", er)
		}
	}
}