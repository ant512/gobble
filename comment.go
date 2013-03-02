package main

import (
	"strings"
	"time"
)

type Comments []*Comment

func (c Comments) Len() int {
	return len(c)
}

func (c Comments) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c Comments) Less(i, j int) bool {
	return c[i].Date.Before(c[j].Date)
}

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
