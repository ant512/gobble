package main

import (
	"time"
	"strings"
	"fmt"
)

type BlogPosts []*BlogPost

func (b BlogPosts) Len() int      {
	return len(b)
}

func (b BlogPosts) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b BlogPosts) Less(i, j int) bool {

	// We use "After" instead of "Before" to get posts in descending order
	return b[i].PublishDate().After(b[j].PublishDate())
}


type BlogPost struct {
	title string
	publishDate time.Time
	tags []string
	body string
}

func (b *BlogPost) SetTitle(s string) {
	b.title = s
}

func (b *BlogPost) Title() string {
	return b.title
}

func (b *BlogPost) SetPublishDate(t time.Time) {
	b.publishDate = t
}

func (b *BlogPost) PublishDate() time.Time {
	return b.publishDate
}

func (b *BlogPost) SetTags(s []string) {
	b.tags = s
}

func (b *BlogPost) Tags() []string {
	return b.tags
}

func (b *BlogPost) SetBody(s string) {
	b.body = s
}

func (b *BlogPost) Body() string {
	return b.body
}

func (b *BlogPost) ContainsTag(tag string) bool {
	for i := range b.Tags() {
		if b.Tags()[i] == tag {
			return true
		}
	}

	return false
}

func (b *BlogPost) Url() string {
	title := strings.ToLower(b.Title())
	title = strings.Replace(title, " ", "-", -1)
	title = strings.Replace(title, ",", "", -1)
	title = strings.Replace(title, "#", "", -1)
	title = strings.Replace(title, ":", "", -1)

	return fmt.Sprintf("%04d/%02d/%02d/%s", b.PublishDate().Year(), b.PublishDate().Month(), b.PublishDate().Day(), title)
}
