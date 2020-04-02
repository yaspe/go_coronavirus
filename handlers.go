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
}


func handleMessage(msg *tgbotapi.Message) HandlerResult {
	if len(msg.From.UserName) == 0 {
		return HandlerResult{"", errors.New("no username - go away"), false}
	}

	chats[msg.From.UserName] = msg.Chat.ID

	parts := strings.Split(msg.Text, " ")
	if parts[0] == "/set_current" {
		if msg.From.UserName != admin {
			return HandlerResult{"", errors.New("you are not admin"), false}
		}
		if len(parts) != 2 {
			return HandlerResult{"", errors.New("args num mismatch"), false}
		}
		e := setCurrent(parts[1])
		if e != nil {
			return HandlerResult{"", e, false}
		}
		return HandlerResult{"ok", nil, false}
	} else if parts[0] == "/shutdown" {
		if msg.From.UserName != admin {
			return HandlerResult{"", errors.New("you are not admin"), false}
		}
		Dump()
		shouldShutdown = true
		return HandlerResult{"ok", nil, false}
	} else if parts[0] == "/switch" {
		if msg.From.UserName != admin {
			return HandlerResult{"", errors.New("you are not admin"), false}
		}
		forceBetsAllowed = !forceBetsAllowed
		return HandlerResult{strconv.FormatBool(forceBetsAllowed), nil, false}
	} else if parts[0] == "/add_winner" {
		if msg.From.UserName != admin {
			return HandlerResult{"", errors.New("you are not admin"), false}
		}
		if len(parts) != 2 {
			return HandlerResult{"", errors.New("args num mismatch"), false}
		}
		winners[parts[1]] ++
		return HandlerResult{"ok", nil, false}
	} else if parts[0] == "/clear" {
		if msg.From.UserName != admin {
			return HandlerResult{"", errors.New("you are not admin"), false}
		}
		bets = make(map[string]int)
		return HandlerResult{"ok", nil, false}
	} else if parts[0] == "/broadcast" {
		if msg.From.UserName != admin {
			return HandlerResult{"", errors.New("you are not admin"), false}
		}
		if len(parts) < 2 {
			return HandlerResult{"", errors.New("args num mismatch"), false}
		}

		return HandlerResult{strings.Join(parts[1:], " "), nil, true}
	} else if parts[0] == "/report" {
		if msg.From.UserName != admin {
			return HandlerResult{"", errors.New("you are not admin"), false}
		}

		if len(parts) != 2 {
			return HandlerResult{"", errors.New("args num mismatch"), false}
		}

		e := setCurrent(parts[1])
		if e != nil {
			return HandlerResult{"", e, false}
		}

		if current == 0 {
			return HandlerResult{"", errors.New("set current"), false}
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
		return HandlerResult{result, nil, true}
	} else if parts[0] == "/dump" {
		if msg.From.UserName != admin {
			return HandlerResult{"", errors.New("you are not admin"), false}
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

		return HandlerResult{result, nil, false}
	} else if parts[0] == "/bet" {
		if len(parts) != 2 {
			return HandlerResult{"", errors.New("неверное числа параметров"), false}
		}

		if !betsAllowed() {
			hours, minutes := mskTime()
			message := fmt.Sprintf("Во избежании нечестной игры, ставки можно делать в интревале %d и %d часов следующего дня по Москве. Дождитесь следующего окна! "+
				"Сейчас %d:%02d", betTimeFrom, betTimeTo, hours, minutes)
			return HandlerResult{message, nil, false}
		}

		var e error
		bet, e := strconv.Atoi(parts[1])
		if e != nil {
			return HandlerResult{"", e, false}
		}

		if current > 0 && bet < current {
			return HandlerResult{"Число заболевших фиксируется с 1го дня и не может уменьшится. Ставка невалидна", nil, false}
		}

		bets[msg.From.UserName] = bet
		return HandlerResult{"Ваша ставка принята!", nil, false}
	} else if parts[0] == "/get" {
		return HandlerResult{strconv.Itoa(current), nil, false}
	} else if parts[0] == "/mybet" {
		return HandlerResult{strconv.Itoa(bets[msg.From.UserName]), nil, false}
	} else if parts[0] == "/github" {
		return HandlerResult{"https://github.com/yaspe/go_coronavirus", nil, false}
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
		return HandlerResult{result, nil, false}
	} else {
		return HandlerResult{help(), nil, false}
	}
}