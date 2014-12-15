package main

import (
	"github.com/go-fsnotify/fsnotify"
	"github.com/russross/blackfriday"
	"log"
	"sync"
)

type Blog struct {
	postPath    string
	commentPath string
	posts       BlogPosts
	tags        map[string]int
	mutex       sync.RWMutex
}

func LoadBlog(postPath, commentPath string) (*Blog, error) {
	b := &Blog{postPath: postPath, commentPath: commentPath}

	b.fetchPosts()
	b.fetchTags()

	return b, nil
}

func (b *Blog) WatchPosts() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				b.fetchPosts()
				b.fetchTags()
				log.Println("Reloading posts due to event:", event)
			case err := <-watcher.Errors:
				log.Println("fswatcher error:", err)
			}
		}
	}()

	err = watcher.Add(b.postPath)
	if err != nil {
		log.Fatal(err)
	}
	<-done
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
	posts, err := LoadBlogPosts(b.postPath, b.commentPath)

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

func convertMarkdownToHtml(markdown *[]byte) string {
	htmlFlags := blackfriday.HTML_USE_SMARTYPANTS
	extensions := blackfriday.EXTENSION_AUTOLINK | blackfriday.EXTENSION_HARD_LINE_BREAK | blackfriday.EXTENSION_FENCED_CODE | blackfriday.EXTENSION_NO_INTRA_EMPHASIS

	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	output := blackfriday.Markdown(*markdown, renderer, extensions)

	return string(output)
}
