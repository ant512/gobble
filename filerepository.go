package main

import (
	"github.com/ant512/gobble/akismet"
	"github.com/howeyc/fsnotify"
	"github.com/russross/blackfriday"
	"html"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type FileRepository struct {
	directory string
	posts     BlogPosts
	tags      map[string]int
	mutex     sync.RWMutex
}

func NewFileRepository(directory string) *FileRepository {

	f := new(FileRepository)
	f.directory = directory

	f.fetchAllPosts()
	f.fetchAllTags()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				f.fetchAllPosts()
				f.fetchAllTags()
				log.Println("Reloading posts due to event:", ev)
			case err := <-watcher.Error:
				log.Println("fswatcher error:", err)
			}
		}
	}()

	err = watcher.Watch(directory)
	if err != nil {
		log.Fatal(err)
	}

	return f
}

func (f *FileRepository) AllTags() map[string]int {
	return f.tags
}

func (f *FileRepository) AllPosts() BlogPosts {
	return f.posts
}

func (f *FileRepository) PostWithUrl(url string) (*BlogPost, error) {

	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.posts.PostWithUrl(url)
}

func (f *FileRepository) PostWithId(id int) (*BlogPost, error) {

	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.posts.PostWithId(id)
}

func (f *FileRepository) PostsWithTag(tag string, start int, count int) (BlogPosts, int) {

	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.posts.PostsWithTag(tag, start, count)
}

func (f *FileRepository) SearchPosts(term string, start int, count int) (BlogPosts, int) {

	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.posts.FilteredPosts(term, start, count)
}

func (f *FileRepository) SaveComment(post *BlogPost, akismetAPIKey, serverAddress, remoteAddress, userAgent, referer, author, email, body string) {

	// TODO: Ensure file name is unique
	isSpam, _ := akismet.IsSpamComment(body, serverAddress, remoteAddress, userAgent, referer, author, email, akismetAPIKey)
	comment := NewComment(html.EscapeString(author), html.EscapeString(email), html.EscapeString(body), isSpam)

	f.mutex.Lock()
	post.Comments = append(post.Comments, comment)
	f.mutex.Unlock()

	postPath := post.Path[:len(post.Path)-3]

	dirname := postPath + string(filepath.Separator) + "comments" + string(filepath.Separator)

	filename := timeToFilename(comment.Date)

	log.Println(dirname + filename)
	os.MkdirAll(dirname, 0775)

	content := comment.String()

	err := ioutil.WriteFile(dirname+filename, []byte(content), 0644)

	if err != nil {
		log.Println(err)
	}
}

func (f *FileRepository) fetchAllPosts() error {

	dirname := f.directory + string(filepath.Separator)

	files, err := ioutil.ReadDir(dirname)

	posts := BlogPosts{}

	for i := range files {

		if files[i].IsDir() {
			continue
		}

		if filepath.Ext(files[i].Name()) != ".md" {
			continue
		}

		post, err := f.fetchPost(dirname + files[i].Name())

		if err != nil {
			return err
		}

		posts = append(posts, post)
	}

	sort.Sort(posts)

	f.mutex.Lock()
	f.posts = posts
	f.mutex.Unlock()

	return err
}

func (f *FileRepository) fetchAllTags() {

	tags := make(map[string]int)

	f.mutex.RLock()

	for i := range f.posts {
		for j := range f.posts[i].Tags {
			tag := strings.ToLower(f.posts[i].Tags[j])
			tag = strings.Replace(tag, "#", "", -1)

			value := tags[tag] + 1
			tags[tag] = value
		}
	}

	f.mutex.RUnlock()

	f.mutex.Lock()
	f.tags = tags
	f.mutex.Unlock()
}

func (f *FileRepository) fetchPost(filename string) (*BlogPost, error) {
	return LoadPost(filename)
}

func (f *FileRepository) fetchComment(filename string) (*Comment, error) {
	comment, err := LoadComment(filename)
	return comment, err
}

func convertMarkdownToHtml(markdown *[]byte) string {
	htmlFlags := blackfriday.HTML_USE_SMARTYPANTS
	extensions := blackfriday.EXTENSION_AUTOLINK | blackfriday.EXTENSION_HARD_LINE_BREAK | blackfriday.EXTENSION_FENCED_CODE | blackfriday.EXTENSION_NO_INTRA_EMPHASIS

	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	output := blackfriday.Markdown(*markdown, renderer, extensions)

	return string(output)
}
