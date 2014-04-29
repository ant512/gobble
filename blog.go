package main

import (
	"github.com/howeyc/fsnotify"
	"log"
	"sync"
)

type Blog struct {
	postPath string
	posts    BlogPosts
	tags     map[string]int
	mutex    sync.RWMutex
}

func LoadBlog(postPath string) (*Blog, error) {
	b := new(Blog)
	b.postPath = postPath

	b.fetchPosts()
	b.fetchTags()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				b.fetchPosts()
				b.fetchTags()
				log.Println("Reloading posts due to event:", ev)
			case err := <-watcher.Error:
				log.Println("fswatcher error:", err)
			}
		}
	}()

	err = watcher.Watch(b.postPath)
	if err != nil {
		log.Fatal(err)
	}

	return b, nil
}

func (b *Blog) AllPosts() BlogPosts {
	return b.posts
}

func (b *Blog) AllTags() map[string]int {
	return b.tags
}

func (b *Blog) PostWithUrl(url string) (*BlogPost, error) {

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return b.posts.PostWithUrl(url)
}

func (b *Blog) PostWithId(id int) (*BlogPost, error) {

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return b.posts.PostWithId(id)
}

func (b *Blog) PostsWithTag(tag string, start int, count int) (BlogPosts, int) {

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return b.posts.PostsWithTag(tag, start, count)
}

func (b *Blog) SearchPosts(term string, start int, count int) (BlogPosts, int) {

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return b.posts.FilteredPosts(term, start, count)
}

func (b *Blog) fetchPosts() {
	posts, err := LoadBlogPosts(b.postPath)

	if err != nil {
		log.Println("Error fetching posts: ", err)
	}

	b.mutex.Lock()
	b.posts = posts
	b.mutex.Unlock()
}

func (b *Blog) fetchTags() {

	tags := make(map[string]int)

	b.mutex.RLock()

	for _, post := range b.posts {
		for _, tag := range post.Tags {
			value := tags[tag] + 1
			tags[tag] = value
		}
	}

	b.mutex.RUnlock()

	b.mutex.Lock()
	b.tags = tags
	b.mutex.Unlock()
}
