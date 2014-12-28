package main

import (
	"errors"
)

const validFilenameExtension = ".md"
const couldNotFindPostErrorMessage = "Could not find post"

type BlogPostFilter func(post *BlogPost, index int, stop *bool) bool
type BlogPosts []*BlogPost

func (b BlogPosts) Len() int {
	return len(b)
}

func (b BlogPosts) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b BlogPosts) Less(i, j int) bool {

	// We use "After" instead of "Before" to get posts in descending order
	return b[i].Metadata.Date.After(b[j].Metadata.Date)
}

func (b BlogPosts) Filter(filter BlogPostFilter) BlogPosts {
	filteredPosts := BlogPosts{}

	stop := false

	for index, post := range b {
		if filter(post, index, &stop) {
			filteredPosts = append(filteredPosts, post)
		}

		if stop {
			break
		}
	}

	return filteredPosts
}

func (b BlogPosts) FilteredPosts(term string, start int, count int) (BlogPosts, int) {
	var filteredPosts BlogPosts

	if len(term) > 0 {
		filteredPosts = b.Filter(func(post *BlogPost, index int, stop *bool) bool {
			return post.ContainsTerm(term)
		})
	} else {
		filteredPosts = b
	}

	if start > len(filteredPosts) {
		return BlogPosts{}, 0
	}

	if start+count > len(filteredPosts) {
		count = len(filteredPosts) - start
	}

	return filteredPosts[start : start+count], len(filteredPosts)
}

func (b BlogPosts) PostWithUrl(url string) (*BlogPost, error) {
	for _, post := range b {
		if post.Url == url {
			return post, nil
		}
	}

	err := errors.New(couldNotFindPostErrorMessage)

	return nil, err
}

func (b BlogPosts) PostsWithTag(tag string, start int, count int) (BlogPosts, int) {

	filteredPosts := b.Filter(func(post *BlogPost, index int, stop *bool) bool {
		return post.ContainsTag(tag)
	})

	if start > len(filteredPosts) {
		return BlogPosts{}, 0
	}

	if start+count > len(filteredPosts) {
		count = len(filteredPosts) - start
	}

	return filteredPosts[start : start+count], len(filteredPosts)
}

func (b BlogPosts) PostWithId(id int) (*BlogPost, error) {
	for _, post := range b {
		if post.Metadata.Id == id {
			return post, nil
		}
	}

	err := errors.New(couldNotFindPostErrorMessage)

	return nil, err
}
