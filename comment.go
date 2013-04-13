package main

import (
	"strings"
	"time"
)

type Comment struct {
	Author string
	Email  string
	Date   time.Time
	Body   string
	IsSpam bool
}

func (c *Comment) ContainsTerm(term string) bool {

	term = strings.ToLower(term)
	terms := strings.Split(term, " ")
	body := strings.ToLower(c.Body)

	for i := range terms {
		if !strings.Contains(body, terms[i]) {
			return false
		}
	}

	return true
}
