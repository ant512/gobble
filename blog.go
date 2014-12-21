package main

import (
	"gopkg.in/fsnotify.v1"
	"io/ioutil"
	"log"
	"path/filepath"
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

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				switch ev.Op {

				// TODO: Need to move the reload/add/remove funcs into this
				// object and handle mutexes.  Or maybe move the mutex into the
				// blogpost object.  That makes more sense.  Maybe this should
				// go in there too, along with all of the paths.
				case fsnotify.Create:
					log.Println("File ", ev.Name, " created")
					b.posts.AddBlogPost(b.postPath, b.commentPath, filepath.Base(ev.Name))
				case fsnotify.Write:
					log.Println("File ", ev.Name, " modified")
					b.posts.ReloadBlogPost(b.postPath, b.commentPath, filepath.Base(ev.Name))
					b.fetchTags()
				case fsnotify.Remove:
					fallthrough
				case fsnotify.Rename:
					log.Println("File ", ev.Name, " deleted")
					b.posts.RemoveBlogPost(filepath.Base(ev.Name))
				}
			case err := <-watcher.Errors:
				log.Println("fswatcher error:", err)
			}
		}
	}()

	err = watcher.Add(b.postPath)
	if err != nil {
		log.Fatal(err)
	}
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
		for _, tag := range post.Metadata.Tags {
			value := tags[tag] + 1
			tags[tag] = value
		}
	}

	b.mutex.RUnlock()

	b.mutex.Lock()
	b.tags = tags
	b.mutex.Unlock()
}

func (b *Blog) fetchChangedPosts() {
	files, err := ioutil.ReadDir(b.postPath)
	log.Println(b.postPath)
	log.Println(err)
	log.Println(files)
}
