package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
)

const validFilenameExtension = ".md"
const couldNotFindPostErrorMessage = "Could not find post"

type BlogPostFilter func(post *BlogPost, index int, stop *bool) bool
type BlogPosts []*BlogPost

func LoadBlogPosts(postPath, commentPath string) (BlogPosts, error) {
	files, err := ioutil.ReadDir(postPath)

	if err != nil {
		return nil, err
	}

	posts := BlogPosts{}

	for _, file := range files {
		if !isValidBlogPostFile(file) {
			continue
		}

		post, err := LoadPost(file.Name(), postPath, commentPath)

		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	sort.Sort(posts)

	return posts, err
}

func (b *BlogPosts) RemoveBlogPost(filename string) error {
	posts := *b

	for i, p := range posts {
		if p.Filename == filename {
			log.Println("Removing post")
			posts = append(posts[:i], posts[i+1:]...)
			break
		}
	}

	*b = posts

	return nil
}

func (b *BlogPosts) AddBlogPost(postPath, commentPath, filename string) error {
	// TODO: If it is valid, load the post, append it, then sort

	fileInfo, err := os.Stat(filepath.Join(postPath, filename))

	if err != nil {
		return err
	}

	if !isValidBlogPostFile(fileInfo) {
		return errors.New("Not a valid blogpost file")
	}

	post, err := LoadPost(fileInfo.Name(), postPath, commentPath)

	if err != nil {
		return err
	}

	posts := *b
	posts = append(posts, post)
	sort.Sort(posts)

	*b = posts

	return nil
}

func (b BlogPosts) ReloadBlogPost(postPath, commentPath, filename string) error {
	fileInfo, err := os.Stat(filepath.Join(postPath, filename))

	if err != nil {
		return err
	}

	if !isValidBlogPostFile(fileInfo) {
		return errors.New("Not a valid blogpost file")
	}

	post, err := LoadPost(fileInfo.Name(), postPath, commentPath)

	if err != nil {
		return err
	}

	replaced := false

	for i, p := range b {
		if p.Filename == filename {
			log.Println("Replacing old post")
			b[i] = post
			replaced = true
			break
		}
	}

	if replaced {
		sort.Sort(b)
	} else {
		err = errors.New("Could not find existing blogpost to replace")
	}

	return err
}

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

func isValidBlogPostFile(fileInfo os.FileInfo) bool {
	if fileInfo.IsDir() {
		return false
	}

	if filepath.Ext(fileInfo.Name()) != validFilenameExtension {
		return false
	}

	return true
}
