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

func handleMessage(msg *tgbotapi.Message) (string, error, bool) {
	if len(msg.From.UserName) == 0 {
		return "", errors.New("no username - go away"), false
	}

	chats[msg.From.UserName] = msg.Chat.ID

	parts := strings.Split(msg.Text, " ")
	if parts[0] == "/set_current" {
		if msg.From.UserName != admin {
			return "", errors.New("you are not admin"), false
		}
		if len(parts) != 2 {
			return "", errors.New("args num mismatch"), false
		}
		var e error
		current, e = strconv.Atoi(parts[1])
		if e != nil {
			return "", e, false
		}
		return "ok", nil, false
	} else if parts[0] == "/shutdown" {
		if msg.From.UserName != admin {
			return "", errors.New("you are not admin"), false
		}
		Dump()
		shouldShutdown = true
		return "ok", nil, false
	} else if parts[0] == "/switch" {
		if msg.From.UserName != admin {
			return "", errors.New("you are not admin"), false
		}
		forceBetable = !forceBetable
		return strconv.FormatBool(forceBetable), nil, false
	} else if parts[0] == "/add_winner" {
		if msg.From.UserName != admin {
			return "", errors.New("you are not admin"), false
		}
		if len(parts) != 2 {
			return "", errors.New("args num mismatch"), false
		}
		winners[parts[1]] ++
		return "ok", nil, false
	} else if parts[0] == "/clear" {
		if msg.From.UserName != admin {
			return "", errors.New("you are not admin"), false
		}
		bets = make(map[string]int)
		return "ok", nil, false
	} else if parts[0] == "/broadcast" {
		if msg.From.UserName != admin {
			return "", errors.New("you are not admin"), false
		}
		if len(parts) < 2 {
			return "", errors.New("args num mismatch"), false
		}

		return strings.Join(parts[1:], " "), nil, true
	} else if parts[0] == "/report" {
		if msg.From.UserName != admin {
			return "", errors.New("you are not admin"), false
		}

		if len(parts) != 1 {
			return "", errors.New("args num mismatch"), false
		}

		if current == 0 {
			return "", errors.New("set current"), false
		}

		top := make(map[int]string)
		for u, b := range bets {
			if b == 0 {
				continue
			}
			diff := int(math.Abs(float64(current - b)))
			top[diff] += "@" + u + " "
		}

		var keys []int
		for k := range top {
			keys = append(keys, k)
		}
		sort.Ints(keys)

		result := "Всего заболевших в России на данный момент: " + strconv.Itoa(current) + "\nпобедители дня(ошибка):\n"
		start := true
		for _, k := range keys {
			result += top[k] + " (" + strconv.Itoa(k) + ")\n"
			if start {
				result += "проиграли:\n"
				start = false
			}
		}

		bets = make(map[string]int)

		forceBetable = true
		return result, nil, true
	} else if parts[0] == "/dump" {
		if msg.From.UserName != admin {
			return "", errors.New("you are not admin"), false
		}

		Dump()

		result := "ставки:\n"
		count := 0
		for u, b := range bets {
			if b == 0 {
				continue
			}
			count++
			result += u + " " + strconv.Itoa(b) + "\n"
		}
		result += "Всего " + strconv.Itoa(count)

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

		return result, nil, false
	} else if parts[0] == "/bet" {
		if len(parts) != 2 {
			return "", errors.New("неверное числа параметров"), false
		}

		if !betable() {
			hours, minutes := mskTime()
			message := fmt.Sprintf("Во избежании нечестной игры, ставки можно делать в интревале %d и %d часов следующего дня по Москве. Дождитесь следующего окна! "+
				"Сейчас %d:%02d", betTimeFrom, betTimeTo, hours, minutes)
			return message, nil, false
		}

		var e error
		bet, e := strconv.Atoi(parts[1])
		if e != nil {
			return "", e, false
		}

		if current > 0 && bet < current {
			return "Число заболевших фиксируется с 1го дня и не может уменьшится. Ставка невалидна", nil, false
		}

		bets[msg.From.UserName] = bet
		return "Ваша ставка принята!", nil, false
	} else if parts[0] == "/get" {
		return strconv.Itoa(current), nil, false
	} else if parts[0] == "/mybet" {
		return strconv.Itoa(bets[msg.From.UserName]), nil, false
	} else if parts[0] == "/github" {
		return "https://github.com/yaspe/go_coronavirus", nil, false
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

		result := "Победители предыдущих дней (количество побед):\n"
		for _, kv := range ss {
			result += kv.Key + " (" + strconv.Itoa(kv.Value) + ")\n"
		}
		return result, nil, false
	} else {
		return help(), nil, false
	}
}