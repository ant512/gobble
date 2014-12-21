package main

import (
	"testing"
	"time"
)

func createPost() *BlogPost {
	p := new(BlogPost)
	p.PostPath = "/test"
	p.Metadata.Id = 2
	p.Metadata.Title = "Test Post"
	p.Metadata.Date = time.Now()
	p.Metadata.DisallowComments = true
	p.Body.Markdown = "Test"
	p.Metadata.Tags = append(p.Metadata.Tags, "test")
	p.Metadata.Tags = append(p.Metadata.Tags, "test1")

	c := NewComment("Joe", "joe@example.com", "This is a test", false)
	p.Comments = append(p.Comments, c)

	c = NewComment("Bob", "bob@example.com", "Another test", true)
	p.Comments = append(p.Comments, c)

	return p
}

func TestNonSpamComments(t *testing.T) {
	p := createPost()

	c := p.NonSpamComments()

	if len(c) != 1 {
		t.Error("Found incorrect number of non-spam comments")
	}
}

func TestContainsTag(t *testing.T) {
	b := createPost()

	if !b.ContainsTag("test") {
		t.Error("Could not find tag")
	}

	if !b.ContainsTag("test1") {
		t.Error("Could not find tag")
	}

	if b.ContainsTag("test2") {
		t.Error("Found missing tag")
	}
}
