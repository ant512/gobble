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
	"strconv"
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

	postPath := post.FilePath[:len(post.FilePath)-3]

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

	post := new(BlogPost)
	post.FilePath = filename

	file, err := ioutil.ReadFile(filename)

	if err != nil {
		return post, err
	}
	
	file = []byte(strings.Replace(string(file), "\r", "", -1))
	file = []byte(extractPostHeader(string(file), post))

	post.Body = convertMarkdownToHtml(&file)

	f.fetchCommentsForPost(post, filename)

	return post, nil
}

func (f *FileRepository) fetchCommentsForPost(post *BlogPost, filename string) {

	dirname := filename[:len(filename)-3] + string(filepath.Separator) + "comments" + string(filepath.Separator)

	files, err := ioutil.ReadDir(dirname)

	if err != nil {
		return
	}

	post.Comments = Comments{}

	for i := range files {

		if files[i].IsDir() {
			continue
		}

		if filepath.Ext(files[i].Name()) != ".md" {
			continue
		}

		comment, err := f.fetchComment(dirname + files[i].Name())

		if err != nil {
			log.Fatal(err)
			return
		}

		post.Comments = append(post.Comments, comment)
	}
}

func (f *FileRepository) fetchComment(filename string) (*Comment, error) {
	comment := new(Comment)

	file, err := ioutil.ReadFile(filename)

	if err != nil {
		return comment, err
	}

	file = []byte(strings.Replace(string(file), "\r", "", -1))
	file = []byte(extractCommentHeader(string(file), comment))

	comment.Body = convertMarkdownToHtml(&file)

	return comment, nil
}

func convertMarkdownToHtml(markdown *[]byte) string {
	htmlFlags := blackfriday.HTML_USE_SMARTYPANTS
	extensions := blackfriday.EXTENSION_AUTOLINK | blackfriday.EXTENSION_HARD_LINE_BREAK | blackfriday.EXTENSION_FENCED_CODE | blackfriday.EXTENSION_NO_INTRA_EMPHASIS

	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	output := blackfriday.Markdown(*markdown, renderer, extensions)

	return string(output)
}

func extractCommentHeader(text string, comment *Comment) string {

	headerSize := parseHeader(text, func(key, value string) {
		switch key {
		case "author":
			comment.Author = value
		case "email":
			comment.Email = value
		case "date":
			comment.Date = stringToTime(value)
		case "spam":
			comment.IsSpam = value == "true"
		default:
		}
	})

	return text[headerSize:]
}

func extractPostHeader(text string, post *BlogPost) string {

	headerSize := parseHeader(text, func(key, value string) {
		switch key {
		case "title":
			post.Title = value
		case "id":
			post.Id, _ = strconv.Atoi(value)
		case "tags":

			tags := strings.Split(value, ",")

			formattedTags := []string{}

			for j := range tags {
				tags[j] = strings.Trim(tags[j], " ")
				tags[j] = strings.Replace(tags[j], " ", "-", -1)
				tags[j] = strings.Replace(tags[j], "/", "-", -1)
				tags[j] = strings.ToLower(tags[j])

				if tags[j] != "" {
					formattedTags = append(formattedTags, tags[j])
				}
			}

			post.Tags = formattedTags
		case "date":
			post.PublishDate = stringToTime(value)
		case "disallowcomments":
			post.DisallowComments = value == "true"
		default:
		}
	})

	return text[headerSize:]
}
