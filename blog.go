package main

import (
	"errors"
	"gopkg.in/fsnotify.v1"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type Blog struct {
	postPath    string
	commentPath string
	posts       BlogPosts
	tags        Tags
	mutex       sync.RWMutex
}

func LoadBlog(postPath, commentPath string, disableWatcher bool) (*Blog, error) {
	b := &Blog{postPath: postPath, commentPath: commentPath}
	b.tags = NewTags()

	err := b.loadBlogPosts()

	if err != nil {
		log.Println("Error fetching posts:", err)
	} else if !disableWatcher {
		b.watchPosts()
	}

	return b, err
}

func (b *Blog) AllPosts() BlogPosts {
	return b.posts
}

func (b *Blog) AllTags() map[string]int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.tags.AllTags()
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

func (b *Blog) loadBlogPosts() error {
	files, err := ioutil.ReadDir(b.postPath)

	if err != nil {
		return err
	}

	posts := BlogPosts{}
	tags := []string{}

	for _, file := range files {
		if !isValidBlogPostFile(file) {
			continue
		}

		post, err := LoadPost(file.Name(), b.postPath, b.commentPath)

		if err != nil {
			return err
		}

		posts = append(posts, post)
		tags = append(tags, post.Metadata.Tags[:]...)
	}

	sort.Sort(posts)

	b.mutex.Lock()
	b.posts = posts
	b.tags.AddTags(tags)
	b.mutex.Unlock()

	return err
}

func (b *Blog) watchPosts() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				switch ev.Op {
				case fsnotify.Create:
					log.Println("File", ev.Name, "created")
					b.addBlogPost(filepath.Base(ev.Name))
				case fsnotify.Write:
					log.Println("File", ev.Name, "modified")
					b.reloadBlogPost(filepath.Base(ev.Name))
				case fsnotify.Remove:
					fallthrough
				case fsnotify.Rename:
					log.Println("File", ev.Name, "deleted")
					b.removeBlogPost(filepath.Base(ev.Name))
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

func (b *Blog) removeBlogPost(filename string) error {
	log.Println("Attempting to remove post with filename", filename)

	posts := b.posts

	var removed *BlogPost = nil

	b.mutex.Lock()
	for i, p := range posts {
		if p.Filename == filename {
			posts = append(posts[:i], posts[i+1:]...)
			removed = p
			break
		}
	}

	b.posts = posts

	var err error = nil

	if removed == nil {
		log.Println("Failed to remove post: post not found")
		err = errors.New("Failed to remove post: post not found")
	} else {
		b.tags.RemoveTags(removed.Metadata.Tags)
		log.Println("Post removed")
	}

	b.mutex.Unlock()

	return err
}

func (b *Blog) addBlogPost(filename string) error {
	log.Println("Attempting to add post with filename" ,filename)

	fileInfo, err := os.Stat(filepath.Join(b.postPath, filename))

	if err != nil {
		return err
	}

	if !isValidBlogPostFile(fileInfo) {
		return errors.New("Not a valid blogpost file")
	}

	post, err := LoadPost(fileInfo.Name(), b.postPath, b.commentPath)

	if err != nil {
		return err
	}

	b.mutex.Lock()
	posts := b.posts
	posts = append(posts, post)
	sort.Sort(posts)

	b.posts = posts
	b.tags.AddTags(post.Metadata.Tags)
	b.mutex.Unlock()

	log.Println("Post added")

	return nil
}

func (b *Blog) reloadBlogPost(filename string) error {
	log.Println("Attempting to reload post with filename", filename)

	fileInfo, err := os.Stat(filepath.Join(b.postPath, filename))

	if err != nil {
		return err
	}

	if !isValidBlogPostFile(fileInfo) {
		return errors.New("Not a valid blogpost file")
	}

	post, err := LoadPost(fileInfo.Name(), b.postPath, b.commentPath)

	if err != nil {
		return err
	}

	var removed *BlogPost = nil

	b.mutex.Lock()
	for i, p := range b.posts {
		if p.Filename == filename {
			b.posts[i] = post
			removed = p
			break
		}
	}

	if removed != nil {
		log.Println("Post reloaded")
		sort.Sort(b.posts)
		b.tags.RemoveTags(removed.Metadata.Tags)
		b.tags.AddTags(post.Metadata.Tags)
	} else {
		log.Println("Failed to reload post: existing post not found")
		err = errors.New("Could not find existing blogpost to replace")
	}

	b.mutex.Unlock()

	return err
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
