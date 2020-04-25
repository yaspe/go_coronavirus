package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type StatData struct {
	Records [][]interface{} `json:"data,omitempty"`
}

func getTodayCurrentFromWiki() (int, error) {
	statData := StatData{}
	stats, err := getStats()
	if err != nil {
		return 0, err
	}

	r := bytes.NewReader([]byte(stats))
	decoder := json.NewDecoder(r)

	err = decoder.Decode(&statData)
	if err != nil {
		return 0, err
	}

	days := statData.Records
	lastDateStats := days[len(days)-1]
	lastDayCases := int(lastDateStats[3].(float64))

	return lastDayCases, err
}

func getStats() (string, error) {
	resp, err := http.Get("https://commons.wikimedia.org/wiki/Data:COVID-19/Cases/RU.tab?action=raw")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}
	result := buf.String() // Does a complete copy of the bytes in the buffer.
	return result, nil
}
