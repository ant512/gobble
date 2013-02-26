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
)

type FileRepository struct {
	directory string
	posts BlogPosts
	tags []string
	mutex sync.RWMutex
}

func NewFileRepository(directory string) *FileRepository {

	f := new(FileRepository)
	f.directory = directory

	go f.update()

	return f
}

func (f *FileRepository) AllTags() []string {
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

	f.mutex.Lock()
	defer f.mutex.Unlock()

	dirname := f.directory + string(filepath.Separator)

	files, err := ioutil.ReadDir(dirname)

	f.posts = BlogPosts{}

	for i := range files {

		if files[i].IsDir() {
			continue
		}

		if filepath.Ext(files[i].Name()) != ".md" {
			continue
		}

		post, err := f.fetchPost(files[i].Name())

		if err != nil {
			return err
		}

		f.posts = append(f.posts, post)
	}

	sort.Sort(f.posts)

	return err
}

func (f *FileRepository) fetchAllTags() {

	f.mutex.Lock()
	defer f.mutex.Unlock()

	// We're using a map to simulate a set
	tagMap := make(map[string]bool)

	for i := range f.posts {
		for j := range f.posts[i].Tags() {
			tagMap[strings.ToLower(f.posts[i].Tags()[j])] = true
		}
	}

	f.tags = []string{}

	for key := range tagMap {
		f.tags = append(f.tags, key)
	}

	sort.Strings(f.tags)
}

func (f *FileRepository) fetchPost(filename string) (*BlogPost, error) {

	post := new(BlogPost)

	dirname := f.directory + string(filepath.Separator)

	file, err := ioutil.ReadFile(dirname + filename)

	if err != nil {
		return post, err
	}

	file = []byte(extractHeader(string(file), post))

	htmlFlags := blackfriday.HTML_USE_SMARTYPANTS
	extensions := blackfriday.EXTENSION_HARD_LINE_BREAK | blackfriday.EXTENSION_FENCED_CODE | blackfriday.EXTENSION_NO_INTRA_EMPHASIS

	renderer := blackfriday.HtmlRenderer(htmlFlags, post.Title(), "")

	output := blackfriday.Markdown(file, renderer, extensions)

	post.SetBody(string(output))

	return post, nil
}

func extractHeader(text string, post *BlogPost) string {

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
