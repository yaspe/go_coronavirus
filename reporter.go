package main

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api" // use develop branch
	"log"
	"math"
	"sort"
	"time"
	"unicode/utf8"
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

		r := report(newCurrent, false)
		if r.Error != nil {
			continue
		}

		for _, chat := range chats {
			msg := tgbotapi.NewMessage(chat, r.Reply)
			er := sendWithMarkup(msg)
			if er != nil {
				fmt.Printf("Could not send message: %s\n", er)
			}
		}

	}
}

func report(newCurrent int, debug bool) *HandlerResult {
	oldCurrent := current
	dailyDiff := newCurrent - oldCurrent
	if !debug {
		current = newCurrent
	}

	if current == 0 {
		return MakeHandlerResultError(errors.New("set current"))
	}

	top := make(map[int][]string)
	for u, b := range bets {
		if b == 0 {
			continue
		}
		diff := int(math.Abs(float64(newCurrent - b)))
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

	if len(keys) > 0 && !debug {
		for _, winner := range top[keys[0]] {
			winners[winner]++
		}
	}

	min, max, avg := minMaxAvgBet()
	result := fmt.Sprintf("\U0001F9A0 Подводим итоги дня!\n"+
		"За прошедние сутки было зафиксировано <b>%s</b> новых заражений, число заболевших достигло <b>%s</b>\n"+
		"Было принято прогнозов: %d\n"+
		"Минимальный:%d Максимальный:%d Средний:%d\n"+
		"<pre>---победители дня|прогноз|дельта:---\n",
		printLargeNumber(dailyDiff), printLargeNumber(newCurrent), betsCount(), min, max, avg)
	longestName := getLongestName()
	start := true
	for _, k := range keys {
		for _, name := range top[k] {
			result += formatName(name)
			spaceLen := longestName - utf8.RuneCountInString(formatName(name))
			for i := 0; i < spaceLen; i++ {
				result += " "
			}
			result += fmt.Sprintf("|%s|+%s\n", printLargeNumber(bets[name]), printLargeNumber(bets[name]-oldCurrent))
		}
		if start {
			result += "---проиграли:---\n"
			start = false
		}
	}
	result += "</pre>"

	if !debug {
		bets = make(map[string]int)
		forceBetsAllowed = true
		go earlyRemind()
		go lateRemind()
		return MakeHandlerResultBroadcast(result)
	}
	return MakeHandlerResultSuccess(result)
}
