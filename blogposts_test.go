package main

import (
	"testing"
	"time"
)

func createTestBlogPosts() BlogPosts {
	b := BlogPosts{}

	bp := new(BlogPost)
	bp.Id = 6
	bp.Title = "Test 1"
	bp.Path = "/test1"
	bp.PublishDate = time.Now()
	bp.Tags = append(bp.Tags, "test")
	bp.Tags = append(bp.Tags, "test1")
	bp.Body = "This is some text"
	bp.DisallowComments = true
	bp.Url = "/test1"
	b = append(b, bp)

	bp = new(BlogPost)
	bp.Id = 8
	bp.Title = "Test 2"
	bp.Path = "/test2"
	bp.PublishDate = time.Now()
	bp.Tags = append(bp.Tags, "test")
	bp.Tags = append(bp.Tags, "test2")
	bp.Body = "This is some more text"
	bp.DisallowComments = false
	bp.Url = "/test2"
	b = append(b, bp)

	return b
}

func TestFilteredPosts(t *testing.T) {
	b := createTestBlogPosts()

	found, _ := b.PostWithId(6)
	if found == nil || found.Title != "Test 1" {
		t.Error("Could not find post by ID")
	}

	found, _ = b.PostWithId(8)
	if found == nil || found.Title != "Test 2" {
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
	if found == nil || found.Title != "Test 1" {
		t.Error("Could not find post with URL")
	}

	found, _ = b.PostWithUrl("/test2")
	if found == nil || found.Title != "Test 2" {
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
		return post.DisallowComments
	})

	if len(bf) > 1 {
		t.Error("Found posts with incorrect disallow comments status")
	}

	if bf[0].Id != 6 {
		t.Error("Found posts with incorrect disallow comments status")
	}
}