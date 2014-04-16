package main

import (
	"errors"
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

func (b BlogPosts) FilteredPosts(term string, start int, count int) (BlogPosts, int) {
	filteredPosts := BlogPosts{}

	if len(term) > 0 {
		for i := range b {
			if b[i].ContainsTerm(term) {
				filteredPosts = append(filteredPosts, b[i])
			}
		}
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
	for i := range b {
		if b[i].Url() == url {
			return b[i], nil
		}
	}

	err := errors.New("Could not find post")

	return nil, err
}

func (b BlogPosts) PostsWithTag(tag string, start int, count int) (BlogPosts, int) {
	filteredPosts := BlogPosts{}

	for i := range b {
		if b[i].ContainsTag(tag) {
			filteredPosts = append(filteredPosts, b[i])
		}
	}

	if start > len(filteredPosts) {
		return BlogPosts{}, 0
	}

	if start+count > len(filteredPosts) {
		count = len(filteredPosts) - start
	}

	return filteredPosts[start : start+count], len(filteredPosts)
}

func (b BlogPosts) PostWithId(id int) (*BlogPost, error) {
	for i := range b {
		if b[i].Id == id {
			return b[i], nil
		}
	}

	err := errors.New("Could not find post")

	return nil, err
}
