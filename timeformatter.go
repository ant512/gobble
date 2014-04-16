package main

import (
	"fmt"
	"strconv"
	"time"
)

func timeToFilename(t time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d_%02d-%02d-%02d.md", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func timeToString(t time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func stringToTime(s string) time.Time {

	year, err := strconv.Atoi(s[:4])

	if err != nil {
		return time.Unix(0, 0)
	}

	month, err := strconv.Atoi(s[5:7])

	if err != nil {
		return time.Unix(0, 0)
	}

	day, err := strconv.Atoi(s[8:10])

	if err != nil {
		return time.Unix(0, 0)
	}

	hour, err := strconv.Atoi(s[11:13])

	if err != nil {
		return time.Unix(0, 0)
	}

	minute, err := strconv.Atoi(s[14:16])

	if err != nil {
		return time.Unix(0, 0)
	}

	seconds, err := strconv.Atoi(s[17:19])

	if err != nil {
		return time.Unix(0, 0)
	}

	location, err := time.LoadLocation("UTC")

	return time.Date(year, time.Month(month), day, hour, minute, seconds, 0, location)
}
