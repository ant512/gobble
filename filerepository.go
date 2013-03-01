package main

import (
	"errors"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"sync"
	"log"
	"fmt"
)

type FileRepository struct {
	directory string
	posts BlogPosts
	tags map[string]int
	mutex sync.RWMutex
}

func NewFileRepository(directory string) *FileRepository {

	f := new(FileRepository)
	f.directory = directory

	go f.update()

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

	for i := range f.posts {
		if f.posts[i].Url() == url {
			return f.posts[i], nil
		}
	}

	err := errors.New("Could not find post")

	return nil, err
}

func (f *FileRepository) PostsWithTag(tag string, start int, count int) (BlogPosts, int) {

	f.mutex.RLock()
	defer f.mutex.RUnlock()

	filteredPosts := BlogPosts{}

	for i := range f.posts {
		if f.posts[i].ContainsTag(tag) {
			filteredPosts = append(filteredPosts, f.posts[i])
		}
	}

	if start > len(filteredPosts) {
		return BlogPosts{}, 0
	}

	if start + count > len(filteredPosts) {
		count = len(filteredPosts) - start
	}

	return filteredPosts[start:start + count], len(filteredPosts)
}

func (f *FileRepository) SearchPosts(term string, start int, count int) (BlogPosts, int) {

	f.mutex.RLock()
	defer f.mutex.RUnlock()

	filteredPosts := BlogPosts{}

	if len(term) > 0 {
		for i := range f.posts {
			if f.posts[i].ContainsTerm(term) {
				filteredPosts = append(filteredPosts, f.posts[i])
			}
		}
	} else {
		filteredPosts = f.posts
	}

	if start > len(filteredPosts) {
		return BlogPosts{}, 0
	}

	if start + count > len(filteredPosts) {
		count = len(filteredPosts) - start
	}

	return filteredPosts[start:start + count], len(filteredPosts)
}

func (f *FileRepository) SaveComment(comment *Comment, post *BlogPost) {

	postPath := post.FilePath()[:len(post.FilePath()) - 3]

	dirname := postPath[:len(postPath) - 3] + string(filepath.Separator) + "comments" + string(filepath.Separator)

	filename := timeToString(comment.Date())

	log.Println(dirname + filename)


}

func (f *FileRepository) update() {

	for {

		start := time.Now()

		f.fetchAllPosts()
		f.fetchAllTags()

		end := time.Now()
		log.Printf("Cached %v posts in %v", len(f.posts), end.Sub(start))

		time.Sleep(10 * time.Minute)
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

	for i := range f.posts {
		for j := range f.posts[i].Tags() {

			value := tags[strings.ToLower(f.posts[i].Tags()[j])] + 1
			tags[strings.ToLower(f.posts[i].Tags()[j])] = value
		}
	}

	f.mutex.Lock()
	f.tags = tags
	f.mutex.Unlock()
}

func (f *FileRepository) fetchPost(filename string) (*BlogPost, error) {

	post := new(BlogPost)
	post.SetFilePath(filename)

	file, err := ioutil.ReadFile(filename)

	if err != nil {
		return post, err
	}

	file = []byte(extractPostHeader(string(file), post))

	htmlFlags := blackfriday.HTML_USE_SMARTYPANTS
	extensions := blackfriday.EXTENSION_AUTOLINK | blackfriday.EXTENSION_HARD_LINE_BREAK | blackfriday.EXTENSION_FENCED_CODE | blackfriday.EXTENSION_NO_INTRA_EMPHASIS

	renderer := blackfriday.HtmlRenderer(htmlFlags, post.Title(), "")

	output := blackfriday.Markdown(file, renderer, extensions)

	post.SetBody(string(output))

	f.fetchCommentsForPost(post, filename)

	return post, nil
}

func (f *FileRepository) fetchCommentsForPost(post *BlogPost, filename string) {

	dirname := filename[:len(filename) - 3] + string(filepath.Separator) + "comments" + string(filepath.Separator)

	files, err := ioutil.ReadDir(dirname)

	if err != nil {
		return
	}

	post.comments = Comments{}

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

		post.comments = append(post.comments, comment)
	}
}

func (f *FileRepository) fetchComment(filename string) (*Comment, error) {
	comment := new(Comment)

	file, err := ioutil.ReadFile(filename)

	if err != nil {
		return comment, err
	}

	file = []byte(extractCommentHeader(string(file), comment))

	htmlFlags := blackfriday.HTML_USE_SMARTYPANTS
	extensions := blackfriday.EXTENSION_AUTOLINK | blackfriday.EXTENSION_HARD_LINE_BREAK | blackfriday.EXTENSION_FENCED_CODE | blackfriday.EXTENSION_NO_INTRA_EMPHASIS

	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	output := blackfriday.Markdown(file, renderer, extensions)

	comment.SetBody(string(output))

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
				comment.SetAuthor(data)
			case "email":
				comment.SetEmail(data)
			case "date":
				comment.SetDate(stringToTime(data))
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
				post.SetTitle(data)
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

				post.SetTags(formattedTags)
			case "date":
				post.SetPublishDate(stringToTime(data))
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

func timeToString(t time.Time) string {
	return fmt.Sprintf("%v-%v-%v_%v-%v-%v", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func stringToTime(s string) time.Time {

	year, err := strconv.Atoi(s[:4])

	if err != nil {
		return time.Unix(0, 0)
	}

	month, err := strconv.Atoi(s[5:7])

	if err != nil {
		return time.Unix(0, 0)
	}

	day, err := strconv.Atoi(s[8:10])

	if err != nil {
		return time.Unix(0, 0)
	}

	hour, err := strconv.Atoi(s[11:13])

	if err != nil {
		return time.Unix(0, 0)
	}

	minute, err := strconv.Atoi(s[14:16])

	if err != nil {
		return time.Unix(0, 0)
	}

	seconds, err := strconv.Atoi(s[17:19])

	if err != nil {
		return time.Unix(0, 0)
	}

	location, err := time.LoadLocation("UTC")

	return time.Date(year, time.Month(month), day, hour, minute, seconds, 0, location)
}
