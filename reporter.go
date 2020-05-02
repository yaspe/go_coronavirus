package main

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"math"
	"sort"
	"time"
)

func reportLoop() {
	for {
		time.Sleep(time.Minute * 30)
		log.Print("reportLoop time")
		if betsAllowed() {
			log.Print("bets allowed nothing to do")
			continue
		}
		newCurrent, e := getTodayCurrentFromWiki()
		if e != nil {
			continue
		}

		if newCurrent <= current {
			log.Print("no new wiki data")
			continue
		}

		r := report(newCurrent)
		if r.Error != nil {
			continue
		}

		for _, chat := range chats {
			msg := tgbotapi.NewMessage(chat, r.Reply)
			_, er := bot.Send(msg)
			if er != nil {
				fmt.Printf("Could not send message: %s\n", er)
			}
		}

	}
}

func report(newCurrent int) *HandlerResult {
	oldCurrent := current
	current = newCurrent
	dailyDiff := current - oldCurrent

	if current == 0 {
		return MakeHandlerResultError(errors.New("set current"))
	}

	top := make(map[int][]string)
	for u, b := range bets {
		if b == 0 {
			continue
		}
		diff := int(math.Abs(float64(current - b)))
		if _, ok := top[diff]; ok {
			top[diff] = append(top[diff], u)
		} else {
			top[diff] = make([]string, 1)
			top[diff][0] = u
		}
	}

	var keys []int
	for k := range top {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	if len(keys) > 0 {
		for _, winner := range top[keys[0]] {
			winners[winner]++
		}
	}

	min, max, avg := minMaxAvgBet()
	result := fmt.Sprintf("Подводим итоги дня!\n"+
		"За прошедние сутки было зафиксировано %s новых заражений, число заболевших достигло %s\n"+
		"Было принято прогнозов: %d\n"+
		"Минимальный: %d\nМаксимальный: %d\nСредний: %d\n\n"+
		"---победители дня(прогноз):---\n",
		printLargeNumber(dailyDiff), printLargeNumber(current), betsCount(), min, max, avg)
	start := true
	for _, k := range keys {
		for i, name := range top[k] {
			if i > 0 {
				result += ", "
			}
			result += formatName(name)
			result += " (" + printLargeNumber(bets[name]) + ")"
		}
		result += "\n"
		if start {
			result += "---проиграли:---\n"
			start = false
		}
	}

	bets = make(map[string]int)

	forceBetsAllowed = true

	go earlyRemind()
	go lateRemind()
	return MakeHandlerResultBroadcast(result)
}
