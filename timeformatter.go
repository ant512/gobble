package main

import (
	"log"
	"time"
)

func timeToFilename(t time.Time) string {
	const layout = "2006-01-02_15-04-05.md"
	return t.Format(layout)
}

func timeToString(t time.Time) string {
	const layout = "2006-01-02 15:04:05"
	return t.Format(layout)
}

func stringToTime(s string) time.Time {
	const layout = "2006-01-02 15:04:05"
	t, err := time.Parse(layout, s)

	if err != nil {
		log.Println(err)
	}

	return t
}
