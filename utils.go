package main

import (
	"fmt"
	"time"
)

func help() string {
	betsInfo := fmt.Sprintf("Сейчас прием ставок открыт и продлится до %d часов утра", betTimeTo)
	if !betable() {
		betsInfo = fmt.Sprintf("Сейчас прием ставок закрыт. Он откроется после объявления результатов предыдущего дня и продлится до %d часов утра", betTimeTo)
	}

	// ★
	h := fmt.Sprintf(
		"Это бот ставок на количество зараженных коронавирусом в России. " +
			"Число заболевших отсчитывается с 1го дня, таким образом каждый день оно должно расти\n" +
			"Подведение итогов в районе 12-15 часов каждого дня, зависит от времени появления новостей\n" +
			"На данный момент заболевших: %d, ставок: %d\n" +
			"Подписано на бота людей: %d\n\n" +
			betsInfo + "\n\n" +
			"/bet <число> : сделать ставку на число заболевших завтра\n" +
			"/mybet : посмотреть свою ставку\n" +
			"/get : узнать число зараженных за прошлый день\n" +
			"/winners : прошлые победители\n" +
			"/github : посмотреть исходники\n",
		current, betsCount(), chatsCount())
	return h
}

func mskTime() (int, int) {
	hours, minutes, _ := time.Now().Clock()
	return hours + 3, minutes
}

func betable() bool {
	hours, _ := mskTime()
	ok := !(hours < betTimeFrom && hours > betTimeTo)
	if ok {
		forceBetable = false
		return true
	}
	if forceBetable {
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
