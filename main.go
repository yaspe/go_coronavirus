package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	admin        = "yaspe"
	adminChatId  = 37129726
	dataFileName = "data.db"
	betTimeFrom  = 15
	betTimeTo    = 7
)

var (
	current          = 0
	shouldShutdown   = false
	bets             = make(map[string]int)
	awaitingBets     = make(map[int64]bool)
	awaitingContact  = make(map[int64]bool)
	chats            = make(map[string]int64)
	winners          = make(map[string]int)
	dataLock         sync.Mutex
	forceBetsAllowed = true
)

var version = "unknown" // -ldflags "-X 'main.version=`git rev-list --all --count`'"
var bot *tgbotapi.BotAPI

func main() {
	Load()
	go reportLoop()
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

		log.Printf("[%s, %d] %s", update.Message.From.UserName, update.Message.Chat.ID, update.Message.Text)

		// pols requires a commit to github.com/go-telegram-bot-api/telegram-bot-api
		// i have a diff and would try to push it
		// see telegram-bot-api.patch
		if update.Message.ForwardFrom != nil && update.Message.ForwardFrom.UserName == admin {
			parts := strings.Split(update.Message.Text, "|")
			if len(parts) >= 3 {
				question := parts[0]
				options := make([]string, 0)
				for i := 1; i < len(parts); i++ {
					options = append(options, parts[i])
				}
				fwd := tgbotapi.NewPoll(adminChatId, question, options)
				ans, er := bot.Send(fwd)
				if er != nil {
					log.Printf("Could not send message: %s", er)
				}
				for _, chat := range chats {
					fwd := tgbotapi.NewForward(chat, chats[admin], ans.MessageID)
					_, er = bot.Send(fwd)
					if er != nil {
						log.Printf("Could not send message: %s", er)
					}
				}
				continue
			}
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
					er := sendWithMarkup(msg)
					if er != nil {
						log.Printf("Could not send message: %s", er)
					}
				}
				continue
			} else if result.ContactMode {
				msg = tgbotapi.NewMessage(adminChatId, result.Reply)
			} else {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, result.Reply)
			}
		}

		er = sendWithMarkup(msg)
		if er != nil {
			log.Printf("Could not send message: %s", er)
		}
	}
}
