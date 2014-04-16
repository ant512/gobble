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

func NewComment(author, email, body string, isSpam bool) *Comment {
	c := new(Comment)
	c.Author = author
	c.Email = email
	c.Date = time.Now()
	c.Body = body
	c.IsSpam = isSpam

	return c
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

func (c *Comment) String() string {
	content := "Author: " + c.Author + "\n"
	content += "Email: " + c.Email + "\n"
	content += "Date: " + timeToString(c.Date) + "\n"

	if c.IsSpam {
		content += "Spam: true\n"
	}

	content += "\n"

	content += c.Body

	return content
}
