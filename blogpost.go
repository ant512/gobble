package main

import (
	"fmt"
	"strings"
	"time"
)

type BlogPosts []*BlogPost

func (b BlogPosts) Len() int {
	return len(b)
}

func (b BlogPosts) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b BlogPosts) Less(i, j int) bool {

	// We use "After" instead of "Before" to get posts in descending order
	return b[i].PublishDate.After(b[j].PublishDate)
}

type BlogPost struct {
	Title       string
	FilePath    string
	PublishDate time.Time
	Tags        []string
	Body        string
	Comments    Comments
}

func (b *BlogPost) NonSpamComments() Comments {
	comments := Comments{}

	for i := range b.Comments {
		if !b.Comments[i].IsSpam {
			comments = append(comments, b.Comments[i])
		}
	}

	return comments
}

func (b *BlogPost) ContainsTag(tag string) bool {
	for i := range b.Tags {
		if b.Tags[i] == strings.ToLower(tag) {
			return true
		}
	}

	return false
}

func (b *BlogPost) ContainsTerm(term string) bool {

	term = strings.ToLower(term)

	if b.ContainsTag(term) {
		return true
	}

	return strings.Contains(strings.ToLower(b.Body), term)
}

func (b *BlogPost) Url() string {
	title := strings.ToLower(b.Title)
	title = strings.Replace(title, " ", "-", -1)
	title = strings.Replace(title, ",", "", -1)
	title = strings.Replace(title, "#", "", -1)
	title = strings.Replace(title, ":", "", -1)
	title = strings.Replace(title, "\"", "", -1)
	title = strings.Replace(title, "?", "", -1)
	title = strings.Replace(title, "/", "", -1)

	return fmt.Sprintf("%04d/%02d/%02d/%s", b.PublishDate.Year(), b.PublishDate.Month(), b.PublishDate.Day(), title)
}
