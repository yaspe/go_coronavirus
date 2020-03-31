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
 	betTimeFrom = 16
 	betTimeTo = 9
)

var (
	current = 0
	shouldShutdown = false
	bets = make(map[string]int)
	chats = make(map[string]int64)
	winners = make(map[string]int)
	dataLock sync.Mutex
	forceBetable = false
)

func main() {
	Load()
	token := os.Args[1]

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		var msg tgbotapi.MessageConfig

		ans, err, all := handleMessage(update.Message)
		if shouldShutdown {
			dataLock.Lock()
			return
		}
		if err != nil {
			log.Printf("Could not process message: %s %s", update.Message.Text, err)
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
			//msg.ReplyToMessageID = update.Message.MessageID
		} else {
			if all {
				for _, chat := range chats {
					log.Printf("Broadcasting message to: %s", chat)
					msg = tgbotapi.NewMessage(chat, ans)
					_, err = bot.Send(msg)
					if err != nil {
						log.Printf("Could not send message: %s", err)
					}
				}
				continue
			} else {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, ans)
			}
		}

		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("Could not send message: %s", err)
		}
	}
}