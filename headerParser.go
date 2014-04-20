package main

import (
	"strings"
)

type ParseHandler func(key, value string)

func ParseHeader(text string, handler ParseHandler) int {
	
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

func stripChars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) < 0 {
			return r
		}
		return -1
	}, str)
}
