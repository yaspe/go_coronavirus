package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func Dump() {
	dataLock.Lock()
	defer dataLock.Unlock()

	_, err := os.Create(dataFileName)
	if err != nil {
		fmt.Println(err)
	}
	db, err := sql.Open("sqlite3", dataFileName)
	defer db.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = db.Exec("CREATE TABLE `bets` (`username` VARCHAR(64) NOT NULL, `chat_id` INTEGER, `bet` INTEGER)")
	if err != nil {
		fmt.Println(err)
		return
	}
	for username, chat := range chats {
		_, err = db.Exec("INSERT INTO `bets` VALUES (?, ?, ?)", username, chat, bets[username])
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	_, err = db.Exec("CREATE TABLE `infected` (`current` INTEGER)")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("INSERT INTO `infected` VALUES (?)", current)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("CREATE TABLE `winners` (`username` VARCHAR(64) NOT NULL, `times` INTEGER)")
	if err != nil {
		fmt.Println(err)
		return
	}

	for username, times := range winners {
		_, err = db.Exec("INSERT INTO `winners` VALUES (?, ?)", username, times)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func Load() {
	dataLock.Lock()
	defer dataLock.Unlock()
	
	db, err := sql.Open("sqlite3", dataFileName)
	defer db.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	rows, err := db.Query("SELECT * FROM `bets`")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			chat int64
			bet   int
			name string
		)
		if err := rows.Scan(&name, &chat, &bet); err != nil {
			fmt.Println(err)
			return
		}
		chats[name] = chat
		bets[name] = bet
	}

	rows, err = db.Query("SELECT * FROM `infected`")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&current); err != nil {
			fmt.Println(err)
			return
		}
	}

	rows, err = db.Query("SELECT * FROM `winners`")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			times   int
			name string
		)
		if err := rows.Scan(&name, &times); err != nil {
			fmt.Println(err)
			return
		}
		winners[name] = times
	}
}