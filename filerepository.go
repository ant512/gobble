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

	// We're using a map to simulate a set
	tags := make(map[string]int)

	f.mutex.RLock()

	for i := range f.posts {
		for j := range f.posts[i].Tags {
			tag := strings.ToLower(f.posts[i].Tags[j])
			tag = stripChars(tag, "#")

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

	file = []byte(stripChars(string(file), "\015"))
	file = []byte(extractPostHeader(string(file), post))

	htmlFlags := blackfriday.HTML_USE_SMARTYPANTS
	extensions := blackfriday.EXTENSION_AUTOLINK | blackfriday.EXTENSION_HARD_LINE_BREAK | blackfriday.EXTENSION_FENCED_CODE | blackfriday.EXTENSION_NO_INTRA_EMPHASIS

	renderer := blackfriday.HtmlRenderer(htmlFlags, post.Title, "")

	output := blackfriday.Markdown(file, renderer, extensions)

	post.Body = string(output)

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

	file = []byte(stripChars(string(file), "\015"))
	file = []byte(extractCommentHeader(string(file), comment))

	htmlFlags := blackfriday.HTML_USE_SMARTYPANTS
	extensions := blackfriday.EXTENSION_AUTOLINK | blackfriday.EXTENSION_HARD_LINE_BREAK | blackfriday.EXTENSION_FENCED_CODE | blackfriday.EXTENSION_NO_INTRA_EMPHASIS

	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	output := blackfriday.Markdown(file, renderer, extensions)

	comment.Body = string(output)

	return comment, nil
}

func extractCommentHeader(text string, comment *Comment) string {

	lines := strings.Split(text, "\n")

	headerSize := 0

	for i := range lines {
		if strings.Contains(lines[i], ":") {
			components := strings.Split(lines[i], ":")

			header := strings.ToLower(strings.Trim(components[0], " "))
			separatorIndex := strings.Index(lines[i], ":") + 1
			data := strings.Trim(lines[i][separatorIndex:], " ")

			switch header {
			case "author":
				comment.Author = data
			case "email":
				comment.Email = data
			case "date":
				comment.Date = stringToTime(data)
			case "spam":
				comment.IsSpam = data == "true"
			default:
				continue
			}

			headerSize += len(lines[i]) + 1
		} else {
			break
		}
	}

	return text[headerSize:]
}

func extractPostHeader(text string, post *BlogPost) string {

	lines := strings.Split(text, "\n")

	headerSize := 0

	for i := range lines {
		if strings.Contains(lines[i], ":") {
			components := strings.Split(lines[i], ":")

			header := strings.ToLower(strings.Trim(components[0], " "))
			separatorIndex := strings.Index(lines[i], ":") + 1
			data := strings.Trim(lines[i][separatorIndex:], " ")

			switch header {
			case "title":
				post.Title = data
			case "id":
				post.Id, _ = strconv.Atoi(data)
			case "tags":

				tags := strings.Split(data, ",")

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
				post.PublishDate = stringToTime(data)
			case "disallowcomments":
				post.DisallowComments = data == "true"
			default:
				continue
			}

			headerSize += len(lines[i]) + 1
		} else {
			break
		}
	}

	return text[headerSize:]
}

func stripChars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) < 0 {
			return r
		}
		return -1
	}, str)
}
