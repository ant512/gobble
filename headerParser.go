package main

import (
	"strings"
)

type parseHandler func(key, value string)

func parseHeader(text string, handler parseHandler) int {

	lines := strings.Split(text, "\n")
	headerSize := 0

	for _, line := range lines {
		if strings.Contains(line, ":") {
			components := strings.Split(line, ":")
			key := strings.ToLower(strings.Trim(components[0], " "))
			separatorIndex := strings.Index(line, ":") + 1
			value := strings.Trim(line[separatorIndex:], " ")

			headerSize += len(line) + 1

			handler(key, value)
		} else {
			break
		}
	}

	return headerSize
}
