package main

import (
	"io/ioutil"
	"strings"
	"time"
)

type Comment struct {
	Author  string
	Email   string
	Date    time.Time
	RawBody string
	Body    string
	IsSpam  bool
}

func LoadComment(path string) (*Comment, error) {
	c := &Comment{}
	file, err := ioutil.ReadFile(path)

	if err != nil {
		return c, err
	}

	file = []byte(strings.Replace(string(file), "\r", "", -1))
	file = []byte(c.extractHeader(string(file)))

	c.RawBody = string(file)
	c.Body = convertMarkdownToHtml(&file)

	return c, nil
}

func NewComment(author, email, body string, isSpam bool) *Comment {

	html := []byte(body)

	c := new(Comment)
	c.Author = author
	c.Email = email
	c.Date = time.Now()
	c.RawBody = body
	c.Body = convertMarkdownToHtml(&html)
	c.IsSpam = isSpam

	return c
}

func (c *Comment) ContainsTerm(term string) bool {

	term = strings.ToLower(term)
	terms := strings.Split(term, " ")
	body := strings.ToLower(c.RawBody)

	for _, item := range terms {
		if !strings.Contains(body, item) {
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

	content += c.RawBody

	return content
}

func (c *Comment) extractHeader(text string) string {

	headerSize := parseHeader(text, func(key, value string) {
		switch key {
		case "author":
			c.Author = value
		case "email":
			c.Email = value
		case "date":
			c.Date = stringToTime(value)
		case "spam":
			c.IsSpam = value == "true"
		default:
		}
	})

	return text[headerSize:]
}
