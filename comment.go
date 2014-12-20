package main

import (
	"log"
	"strings"
	"time"
)

type CommentMetadata struct {
	Author  string
	Email   string
	Date    time.Time
	IsSpam  bool
}

type CommentBody struct {
	Markdown string
	HTML    string
}

type Comment struct {
	Metadata  CommentMetadata
	Body CommentBody
}

func LoadComment(path string) (*Comment, error) {
	c := &Comment{}

	err := loadBlogFile(path, func(key, value string) {
		switch key {
		case "author":
			c.Metadata.Author = value
		case "email":
			c.Metadata.Email = value
		case "date":
			c.Metadata.Date = stringToTime(value)
		case "spam":
			c.Metadata.IsSpam = value == "true"
		default:
		}
	}, func(text string) {
		bytes := []byte(text)

		c.Body.Markdown = text
		c.Body.HTML = convertMarkdownToHtml(&bytes)
	})

	if (err != nil) {
		log.Println(err)
	}

	return c, err
}

func NewComment(author, email, body string, isSpam bool) *Comment {

	html := []byte(body)

	c := new(Comment)
	c.Metadata.Author = author
	c.Metadata.Email = email
	c.Metadata.Date = time.Now()
	c.Metadata.IsSpam = isSpam
	c.Body.Markdown = body
	c.Body.HTML = convertMarkdownToHtml(&html)

	return c
}

func (b *CommentBody) ContainsTerm(term string) bool {
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

func (m *CommentMetadata) ContainsTerm(term string) bool {
	term = strings.ToLower(term)
	terms := strings.Split(term, " ")
	body := strings.ToLower(m.Author)

	for _, item := range terms {
		if !strings.Contains(body, item) {
			return false
		}
	}

	return true
}

func (c *Comment) ContainsTerm(term string) bool {
	return c.Metadata.ContainsTerm(term) || c.Body.ContainsTerm(term)
}

func (c *Comment) String() string {
	content := c.Metadata.String()
	content += "\n"
	content += c.Body.String()

	return content
}

func (b *CommentBody) String() string {
	return b.Markdown
}

func (m *CommentMetadata) String() string {
	content := "Author: " + m.Author + "\n"
	content += "Email: " + m.Email + "\n"
	content += "Date: " + timeToString(m.Date) + "\n"

	if m.IsSpam {
		content += "Spam: true\n"
	}

	return content
}
