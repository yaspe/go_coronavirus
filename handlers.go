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


func setCurrent(s string) error {
	var e error
	current, e = strconv.Atoi(s)
	return e
}


type HandlerResult struct {
	Reply string
	Error error
	BroadCast bool
	RemindMode bool
}

func MakeHandlerResultSuccess(reply string) *HandlerResult {
	return &HandlerResult{reply, nil, false, false}
}

func MakeHandlerResultBroadcast(reply string) *HandlerResult {
	return &HandlerResult{reply, nil, true, false}
}

func MakeHandlerResultRemind(reply string) *HandlerResult {
	return &HandlerResult{reply, nil, true, true}
}

func MakeHandlerResultError(e error) *HandlerResult {
	return &HandlerResult{"", e, false, false}
}

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
		winners[parts[1]] ++
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

		e := setCurrent(parts[1])
		if e != nil {
			return MakeHandlerResultError(e)
		}

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
				winners[winner] ++
			}
		}

		result := "Всего заболевших в России на данный момент: " + strconv.Itoa(current) + "\n---победители дня(ошибка):---\n"
		start := true
		for _, k := range keys {
			for i, name := range(top[k]) {
				if i > 0 {
					result += ", "
				}
				result += formatName(name)
			}
			result += " (" + strconv.Itoa(k) + ")\n"
			if start {
				result += "---проиграли:---\n"
				start = false
			}
		}

		bets = make(map[string]int)

		forceBetsAllowed = true
		return MakeHandlerResultBroadcast(result)
	} else if parts[0] == "/dump" {
		if msg.From.UserName != admin {
			return MakeHandlerResultError(errors.New("you are not admin"))
		}

		Dump()

		result := "ставки:\n"
		count := 0
		for u, b := range bets {
			if b == 0 {
				continue
			}
			count++
			result += formatName(u) + " " + strconv.Itoa(b) + "\n"
		}
		result += "Всего " + strconv.Itoa(count)

		if len(parts) == 2 {
			result += "\n\nчаты:\n"
			count = 0
			for u, c := range chats {
				if c == 0 {
					continue
				}
				count++
				result += u + " " + strconv.Itoa(int(c)) + "\n"
			}
			result += "Всего " + strconv.Itoa(count)
		}

		return MakeHandlerResultSuccess(result)
	} else if parts[0] == "/bet" {
		if len(parts) != 2 {
			return MakeHandlerResultError(errors.New("неверное числа параметров"))
		}

		if !betsAllowed() {
			hours, minutes := mskTime()
			message := fmt.Sprintf("Во избежании нечестной игры, ставки можно делать в интревале %d и %d часов следующего дня по Москве. Дождитесь следующего окна! "+
				"Сейчас %d:%02d", betTimeFrom, betTimeTo, hours, minutes)
			return MakeHandlerResultSuccess(message)
		}

		var e error
		bet, e := strconv.Atoi(parts[1])
		if e != nil {
			return MakeHandlerResultError(e)
		}

		if current > 0 && bet < current {
			return MakeHandlerResultError(errors.New("Число заболевших фиксируется с 1го дня и не может уменьшится. Ставка невалидна"))
		}

		bets[msg.From.UserName] = bet
		return MakeHandlerResultSuccess("Ваша ставка принята!")
	} else if parts[0] == "/get" {
		return MakeHandlerResultSuccess(strconv.Itoa(current))
	} else if parts[0] == "/mybet" {
		return MakeHandlerResultSuccess(strconv.Itoa(bets[msg.From.UserName]))
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
	} else {
		return MakeHandlerResultSuccess(help())
	}
}