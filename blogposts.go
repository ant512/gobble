package main

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"sort"
)

type BlogPostFilter func(post *BlogPost, index int, stop *bool) bool
type BlogPosts []*BlogPost

func LoadBlogPosts(path string) (BlogPosts, error) {
	dirname := path + string(filepath.Separator)

	files, err := ioutil.ReadDir(dirname)

	posts := BlogPosts{}

	for i := range files {

		if files[i].IsDir() {
			continue
		}

		if filepath.Ext(files[i].Name()) != ".md" {
			continue
		}

		post, err := LoadPost(dirname + files[i].Name())

		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	sort.Sort(posts)

	return posts, err
}

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

func (b BlogPosts) Filter(filter BlogPostFilter) BlogPosts {
	filteredPosts := BlogPosts{}

	stop := new(bool)
	*stop = false

	for index, post := range b {
		if filter(post, index, stop) {
			filteredPosts = append(filteredPosts, post)
		}

		if *stop {
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

	err := errors.New("Could not find post")

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
		if post.Id == id {
			return post, nil
		}
	}

	err := errors.New("Could not find post")

	return nil, err
}
