package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"math"
	"strconv"
	"time"
)

func help() string {
	betsInfo := fmt.Sprintf("Сейчас прием прогнозов открыт и продлится до %d часов утра по Москве", betTimeTo)
	if !betsAllowed() {
		betsInfo = fmt.Sprintf("Сейчас прием прогнозов закрыт. Он откроется после объявления результатов предыдущего дня и продлится до %d часов утра по Москве", betTimeTo)
	}

	// ★
	h := fmt.Sprintf(
		"Это бот прогнозов на количество зараженных коронавирусом в России. "+
			"Число заболевших отсчитывается с 1го дня, таким образом каждый день оно должно расти\n"+
			"Подведение итогов в районе 11-15 часов каждого дня, зависит от времени появления новостей\n"+
			"На данный момент заболевших: %d, прогнозов: %d\n"+
			"Подписано на бота: %d\n\n"+
			betsInfo+"\n\n"+
			"/bet <число> : сделать прогноз на число заболевших завтра. Если число меньше количества заболевших сейчас, оно трактуется как прирост. "+
			"Если больше - как итоговое число заболевших\n"+
			"/mybet : посмотреть свой прогноз\n"+
			"/get : узнать число зараженных за прошлый день\n"+
			"/winners : прошлые победители\n"+
			"/help : посмотреть это сообщение\n"+
			"Версия %s",
		current, betsCount(), chatsCount(), version)
	return h
}

func mskTime() (int, int) {
	hours, minutes, _ := time.Now().Clock()
	return hours + 3, minutes
}

func betsAllowed() bool {
	hours, _ := mskTime()
	ok := !(hours < betTimeFrom && hours > betTimeTo)
	if ok {
		forceBetsAllowed = false
		return true
	}
	if forceBetsAllowed {
		return true
	}
	return false
}

func betsCount() int {
	c := 0
	for _, b := range bets {
		if b == 0 {
			continue
		}
		c++
	}
	return c
}

func chatsCount() int {
	c := 0
	for _, v := range chats {
		if v == 0 {
			continue
		}
		c++
	}
	return c
}

func formatName(name string) string {
	ret := "@" + name
	if times, ok := winners[name]; ok {
		for i := 0; i < times; i++ {
			ret += "★"
		}
	}
	return ret
}

func setCurrent(s string) error {
	var e error
	current, e = strconv.Atoi(s)
	return e
}

func remindLater(d time.Duration, msg string) {
	time.Sleep(d)
	for username, chat := range chats {
		if val, ok := bets[username]; ok && val > 0 {
			continue
		}
		msg := tgbotapi.NewMessage(chat, msg)
		_, er := bot.Send(msg)
		if er != nil {
			fmt.Printf("Could not send message: %s\n", er)
		}
	}
}

func earlyRemind() {
	text := fmt.Sprintf("Прием прогнозов на сегодня открыт и продлится до %d часов утра по Москве!\n", betTimeTo)
	remindLater(5*time.Minute, text)
}

func lateRemind() {
	remindLater(
		5*time.Hour,
		fmt.Sprintf("Напоминаем, что прием прогнозов открыт!\n"+
			"Заболевших вчера - %d\n"+
			"Прием прогнозов продлится до %d часов утра по Москве\n"+
			"Сделать прозноз: /bet <число>",
			current, betTimeTo))
}

func dumpLoop() {
	for {
		time.Sleep(15 * time.Minute)
		Dump()
	}
}

func betInfo(inc, total int) string {
	return fmt.Sprintf("Ваш прогноз: завтра число заболевших прирастет на %d и составит %d", inc, total)
}

func minMaxAvgBet() (int, int, int) {
	var min, max, sum, num int
	min = math.MaxInt32
	for _, val := range bets {
		if val == 0 {
			continue
		}
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
		sum += val
		num++
	}
	return min, max, sum / num
}

func calcBet(bet int) (int, int) {
	var inc, total int
	if bet < current {
		inc = bet
		total = bet + current
	} else {
		total = bet
		inc = total - current
	}
	return inc, total
}
