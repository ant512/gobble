package main

import (
	"strings"
)

type BlogItemBody struct {
	Markdown string
	HTML     string
}

func (b *BlogItemBody) ContainsTerm(term string) bool {
	term = strings.ToLower(term)
	terms := strings.Split(term, " ")
	body := strings.ToLower(b.Markdown)

	for _, item := range terms {
		if !strings.Contains(body, item) {
			return false
		}
	}

	return true
}

func (b *BlogItemBody) String() string {
	return b.Markdown
}
