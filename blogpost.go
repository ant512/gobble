package main

import (
	"fmt"
	"strings"
	"time"
)

type BlogPost struct {
	Title            string
	Id               int
	FilePath         string
	PublishDate      time.Time
	Tags             []string
	Body             string
	Comments         Comments
	DisallowComments bool
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

	if b.Comments.ContainsTerm(term) {
		return true
	}

	terms := strings.Split(term, " ")
	body := strings.ToLower(b.Body)
	title := strings.ToLower(b.Title)

	for i := range terms {
		if !strings.Contains(body, terms[i]) && !strings.Contains(title, terms[i]) {
			return false
		}
	}

	return true
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

func (b *BlogPost) AllowsComments() bool {
	if b.DisallowComments {
		return false
	}

	if SharedConfig.CommentsOpenForDays == 0 {
		return true
	}

	var closeDate = b.PublishDate.Add(time.Hour * 24 * time.Duration(SharedConfig.CommentsOpenForDays))

	return time.Now().Before(closeDate)
}
