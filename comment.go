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
	return strings.Contains(strings.ToLower(c.Body), term)
}
