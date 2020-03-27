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
			"Ставки принимаются с %d до %d часов по Москве. Сейчас %d:%02d\n" +
			"Подведение итогов в районе 18 часов каждого дня\n" +
			"На данный момент заболевших: %d\n\n" +
			"/bet <число> : сделать ставку на число заболевших завтра\n" +
			"/mybet : посмотреть свою ставку\n" +
			"/get : узнать число зараженных за прошлый день",
		betTimeFrom, betTimeTo, hours, minutes, current)
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
