package main

import (
	"fmt"
	"time"
)

func help() string {
	hours, minutes := msk_time()
	h := fmt.Sprintf(
		"Это бот ставок на количество зараженных коронавирусом в России. " +
			"Число заболевших отсчитывается с 1го дня, таким образом каждый день оно должно расти\n" +
			"Ставки принимаются с %d до %d часов по Москве (или если явно включены). Сейчас %d:%02d. Ставок: %d\n" +
			"Подведение итогов в районе 12-15 часов каждого дня, зависит от времени появления новостей\n" +
			"На данный момент заболевших: %d\n\n" +
			"/bet <число> : сделать ставку на число заболевших завтра\n" +
			"/mybet : посмотреть свою ставку\n" +
			"/get : узнать число зараженных за прошлый день",
		betTimeFrom, betTimeTo, hours, minutes, betsCount(), current)
	return h
}

func msk_time() (int, int) {
	hours, minutes, _ := time.Now().Clock()
	return hours + 3, minutes
}

func betable() bool {
	hours, _ := msk_time()
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
