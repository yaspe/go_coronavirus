package main

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"math"
	"sort"
	"strconv"
	"strings"
)

func handleMessage(msg *tgbotapi.Message) *HandlerResult {
	if len(msg.From.UserName) == 0 {
		return MakeHandlerResultError(errors.New("no username - go away"))
	}

	chats[msg.From.UserName] = msg.Chat.ID

	parts := strings.Split(msg.Text, " ")

	if parts[0] == "/set_current" {
		if msg.From.UserName != admin {
			return MakeHandlerResultError(errors.New("you are not admin"))
		}
		if len(parts) != 2 {
			return MakeHandlerResultError(errors.New("args num mismatch"))
		}
		e := setCurrent(parts[1])
		if e != nil {
			return MakeHandlerResultError(e)
		}
		return MakeHandlerResultSuccess("ok")
	} else if parts[0] == "/shutdown" {
		if msg.From.UserName != admin {
			return MakeHandlerResultError(errors.New("you are not admin"))
		}
		Dump()
		shouldShutdown = true
		return MakeHandlerResultSuccess("ok")
	} else if parts[0] == "/switch" {
		if msg.From.UserName != admin {
			return MakeHandlerResultError(errors.New("you are not admin"))
		}
		forceBetsAllowed = !forceBetsAllowed
		return MakeHandlerResultSuccess(strconv.FormatBool(forceBetsAllowed))
	} else if parts[0] == "/add_winner" {
		if msg.From.UserName != admin {
			return MakeHandlerResultError(errors.New("you are not admin"))
		}
		if len(parts) != 2 {
			return MakeHandlerResultError(errors.New("args num mismatch"))
		}
		winners[parts[1]]++
		return MakeHandlerResultSuccess("ok")
	} else if parts[0] == "/clear" {
		if msg.From.UserName != admin {
			return MakeHandlerResultError(errors.New("you are not admin"))
		}
		bets = make(map[string]int)
		return MakeHandlerResultSuccess("ok")
	} else if parts[0] == "/broadcast" {
		if msg.From.UserName != admin {
			return MakeHandlerResultError(errors.New("you are not admin"))
		}
		if len(parts) < 2 {
			return MakeHandlerResultError(errors.New("args num mismatch"))
		}

		return MakeHandlerResultBroadcast(strings.Join(parts[1:], " "))
	} else if parts[0] == "/remind" {
		if msg.From.UserName != admin {
			return MakeHandlerResultError(errors.New("you are not admin"))
		}
		if len(parts) < 2 {
			return MakeHandlerResultError(errors.New("args num mismatch"))
		}

		return MakeHandlerResultRemind(strings.Join(parts[1:], " "))
	} else if parts[0] == "/report" {
		if msg.From.UserName != admin {
			return MakeHandlerResultError(errors.New("you are not admin"))
		}

		if len(parts) != 2 {
			return MakeHandlerResultError(errors.New("args num mismatch"))
		}

		oldCurrent := current
		e := setCurrent(parts[1])
		if e != nil {
			return MakeHandlerResultError(e)
		}
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
			"За прошедние сутки было зафиксировано %d новых заражений, число заболевших достигло %d\n"+
			"Было принято прогнозов: %d\n"+
			"Минимальный: %d\nМаксимальный: %d\nСредний: %d\n\n"+
			"---победители дня(прогноз):---\n",
			dailyDiff, current, betsCount(), min, max, avg)
		start := true
		for _, k := range keys {
			for i, name := range top[k] {
				if i > 0 {
					result += ", "
				}
				result += formatName(name)
				result += " (" + strconv.Itoa(bets[name]) + ")"
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
	} else if parts[0] == "/dump" {
		if msg.From.UserName != admin {
			return MakeHandlerResultError(errors.New("you are not admin"))
		}
		Dump()
		return MakeHandlerResultSuccess("ok")
	} else if parts[0] == "/bet" || (len(parts) == 1 && awaitingBets[msg.Chat.ID]) {
		if !betsAllowed() {
			hours, minutes := mskTime()
			message := fmt.Sprintf("Во избежании нечестной игры, прознозы можно делать в интревале %d и %d часов следующего дня по Москве. Дождитесь следующего окна! "+
				"Сейчас %d:%02d", betTimeFrom, betTimeTo, hours, minutes)
			return MakeHandlerResultSuccess(message)
		}

		var bet int
		var e error

		if len(parts) > 2 {
			return MakeHandlerResultError(errors.New("неверное числа параметров"))
		} else if len(parts) == 1 {
			if awaitingBets[msg.Chat.ID] {
				awaitingBets[msg.Chat.ID] = false
				bet, e = strconv.Atoi(parts[0])
				if e != nil {
					return MakeHandlerResultError(e)
				}
			} else {
				awaitingBets[msg.Chat.ID] = true
				return MakeHandlerResultSuccess("Пришлите вас прогноз - сколько будет зараженных завтра?")
			}
		} else { // == 2
			bet, e = strconv.Atoi(parts[1])
			if e != nil {
				return MakeHandlerResultError(e)
			}
		}

		var inc, total int = calcBet(bet)
		bets[msg.From.UserName] = total
		return MakeHandlerResultSuccess(betInfo(inc, total))
	} else if parts[0] == "/get" {
		return MakeHandlerResultSuccess(fmt.Sprintf("На данный момент заболевших: %d", current))
	} else if parts[0] == "/mybet" {
		total := bets[msg.From.UserName]
		if total == 0 {
			return MakeHandlerResultSuccess("Вы еще не делали прозноз")
		}
		inc := total - current
		return MakeHandlerResultSuccess(betInfo(inc, total))
	} else if parts[0] == "/github" {
		return MakeHandlerResultSuccess("https://github.com/yaspe/go_coronavirus")
	} else if parts[0] == "/winners" {

		type kv struct {
			Key   string
			Value int
		}

		var ss []kv
		for k, v := range winners {
			ss = append(ss, kv{k, v})
		}

		sort.Slice(ss, func(i, j int) bool {
			return ss[i].Value > ss[j].Value
		})

		result := "Победители предыдущих дней (каждая звездочка - победа):\n"
		for _, kv := range ss {
			result += formatName(kv.Key) + "\n"
		}
		return MakeHandlerResultSuccess(result)
	} else if parts[0] == "/list_bets" {
		result := "прознозы:\n"
		count := 0
		for u, b := range bets {
			if b == 0 {
				continue
			}
			count++
			result += formatName(u) + " " + strconv.Itoa(b) + "\n"
		}
		result += "Всего " + strconv.Itoa(count)
		return MakeHandlerResultSuccess(result)
	} else {
		return MakeHandlerResultSuccess(help())
	}
}
