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
	return c[i].Date().Before(c[j].Date())
}

type Comment struct {
	author      string
	email       string
	date        time.Time
	body        string
}

func (c *Comment) SetAuthor(s string) {
	c.author = s
}

func (c *Comment) Author() string {
	return c.author
}

func (c *Comment) SetEmail(s string) {
	c.email = s
}

func (c *Comment) Email() string {
	return c.email
}

func (c *Comment) SetDate(t time.Time) {
	c.date = t
}

func (c *Comment) Date() time.Time {
	return c.date
}

func (c *Comment) SetBody(s string) {
	c.body = s
}

func (c *Comment) Body() string {
	return c.body
}

func (c *Comment) ContainsTerm(term string) bool {
	return strings.Contains(strings.ToLower(c.body), term)
}
