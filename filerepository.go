package main

import (
	"errors"
	"fmt"
	"github.com/ant512/gobble/akismet"
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
	"time"
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

func (f *FileRepository) PostWithId(id int) (*BlogPost, error) {

	f.mutex.RLock()
	defer f.mutex.RUnlock()

	for i := range f.posts {
		if f.posts[i].Id == id {
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

	if start+count > len(filteredPosts) {
		count = len(filteredPosts) - start
	}

	return filteredPosts[start : start+count], len(filteredPosts)
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

	if start+count > len(filteredPosts) {
		count = len(filteredPosts) - start
	}

	return filteredPosts[start : start+count], len(filteredPosts)
}

func (f *FileRepository) SaveComment(post *BlogPost, akismetAPIKey, serverAddress, remoteAddress, userAgent, referer, author, email, body string) {

	// TODO: Ensure file name is unique
	isSpam, _ := akismet.IsSpamComment(body, serverAddress, remoteAddress, userAgent, referer, author, email, akismetAPIKey)

	comment := new(Comment)

	comment.Author = html.EscapeString(author)
	comment.Email = html.EscapeString(email)
	comment.Date = time.Now()
	comment.Body = html.EscapeString(body)
	comment.IsSpam = isSpam

	f.mutex.Lock()
	post.Comments = append(post.Comments, comment)
	f.mutex.Unlock()

	postPath := post.FilePath[:len(post.FilePath)-3]

	dirname := postPath + string(filepath.Separator) + "comments" + string(filepath.Separator)

	filename := timeToFilename(comment.Date)

	log.Println(dirname + filename)
	os.MkdirAll(dirname, 0775)

	content := "Author: " + comment.Author + "\n"
	content += "Email: " + comment.Email + "\n"
	content += "Date: " + timeToString(comment.Date) + "\n"

	if isSpam {
		content += "Spam: true\n"
	}

	content += "\n"

	content += comment.Body

	err := ioutil.WriteFile(dirname+filename, []byte(content), 0644)

	if err != nil {
		log.Println(err)
	}
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

	f.mutex.RLock()

	for i := range f.posts {
		for j := range f.posts[i].Tags {

			value := tags[strings.ToLower(f.posts[i].Tags[j])] + 1
			tags[strings.ToLower(f.posts[i].Tags[j])] = value
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

func timeToFilename(t time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d_%02d-%02d-%02d.md", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func timeToString(t time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
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
