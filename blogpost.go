package main

import (
	"time"
)

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
