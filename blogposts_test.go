package main

import (
	"testing"
	"time"
)

func createTestBlogPosts() BlogPosts {
	b := BlogPosts{}

	bp := new(BlogPost)
	bp.Url = "/test1"
	bp.Metadata.Id = 6
	bp.Metadata.Title = "Test 1"
	bp.PostPath = "/test1"
	bp.Metadata.Date = time.Now()
	bp.Metadata.Tags = append(bp.Metadata.Tags, "test")
	bp.Metadata.Tags = append(bp.Metadata.Tags, "test1")
	bp.Body.Markdown = "This is some text"
	bp.Metadata.DisallowComments = true
	b = append(b, bp)

	bp = new(BlogPost)
	bp.Metadata.Id = 8
	bp.Metadata.Title = "Test 2"
	bp.PostPath = "/test2"
	bp.Metadata.Date = time.Now()
	bp.Metadata.Tags = append(bp.Metadata.Tags, "test")
	bp.Metadata.Tags = append(bp.Metadata.Tags, "test2")
	bp.Body.Markdown = "This is some more text"
	bp.Metadata.DisallowComments = false
	bp.Url = "/test2"
	b = append(b, bp)

	return b
}

func TestFilteredPosts(t *testing.T) {
	b := createTestBlogPosts()

	found, _ := b.PostWithId(6)
	if found == nil || found.Metadata.Title != "Test 1" {
		t.Error("Could not find post by ID")
	}

	found, _ = b.PostWithId(8)
	if found == nil || found.Metadata.Title != "Test 2" {
		t.Error("Could not find post by ID")
	}

	found, err := b.PostWithId(7)
	if err == nil {
		t.Error("Found post with missing ID")
	}
}

func TestPostWithUrl(t *testing.T) {
	b := createTestBlogPosts()

	found, _ := b.PostWithUrl("/test1")
	if found == nil || found.Metadata.Title != "Test 1" {
		t.Error("Could not find post with URL")
	}

	found, _ = b.PostWithUrl("/test2")
	if found == nil || found.Metadata.Title != "Test 2" {
		t.Error("Could not find post with URL")
	}

	found, err := b.PostWithUrl("/test3")
	if err == nil {
		t.Error("Found post with missing URL")
	}
}

func TestPostsWithTag(t *testing.T) {
	b := createTestBlogPosts()

	_, count := b.PostsWithTag("test", 0, len(b))
	if count != 2 {
		t.Error("Could not find posts with tag")
	}

	_, count = b.PostsWithTag("test1", 0, len(b))
	if count != 1 {
		t.Error("Could not find posts with tag")
	}

	_, count = b.PostsWithTag("tes", 0, len(b))
	if count != 0 {
		t.Error("Found posts with missing tag")
	}
}

func TestFilter(t *testing.T) {
	b := createTestBlogPosts()

	bf := b.Filter(func(post *BlogPost, index int, stop *bool) bool {
		return post.Metadata.DisallowComments
	})

	if len(bf) > 1 {
		t.Error("Found posts with incorrect disallow comments status")
	}

	if bf[0].Metadata.Id != 6 {
		t.Error("Found posts with incorrect disallow comments status")
	}
}
