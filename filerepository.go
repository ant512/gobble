package main

import (
	"github.com/russross/blackfriday"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
	"strconv"
	"sort"
	"errors"
)

type FileRepository struct {
	postDirectory string
}

func (f *FileRepository) FetchAllTags() ([]string, error) {
	posts, err := f.FetchAllPosts()

	// We're using a map to simulate a set
	tagMap := make(map[string]bool)

	for i := range posts {
		for j := range posts[i].Tags() {
			tagMap[posts[i].Tags()[j]] = true
		}
	}

	tags := []string{}

	for key := range tagMap {
		tags = append(tags, key)
	}

	sort.Strings(tags)

	return tags, err
}

func (f *FileRepository) FetchPostWithUrl(url string) (*BlogPost, error) {
	posts, err := f.FetchAllPosts()

	if err != nil {
		return nil, err
	}

	for i := range posts {
		if posts[i].Url() == url {
			return posts[i], err
		}
	}

	err = errors.New("Could not find post")

	return nil, err
}

func (f *FileRepository) FetchPostsWithTag(tag string) ([]*BlogPost, error) {
	posts, err := f.FetchAllPosts()

	if err != nil {
		return nil, err
	}

	filteredPosts := []*BlogPost{}

	for i := range posts {
		if posts[i].ContainsTag(tag) {
			filteredPosts = append(filteredPosts, posts[i])
		}
	}

	return filteredPosts, err
}

func (f *FileRepository) PostDirectory() string {
	return f.postDirectory
}

func (f *FileRepository) SetPostDirectory(s string) {
	f.postDirectory = s
}

func (f *FileRepository) FetchPostsInRange(start, end int) (BlogPosts, error) {
	posts, err := f.FetchAllPosts()

	return posts[start:end], err
}

func (f *FileRepository) FetchAllPosts() (BlogPosts, error) {

	dirname := f.postDirectory + string(filepath.Separator)

	files, err := ioutil.ReadDir(dirname)

	posts := BlogPosts{}

	for i := range files {

		if files[i].IsDir() {
			continue
		}

		if filepath.Ext(files[i].Name()) != ".md" {
			continue
		}

		post, err := f.FetchPost(files[i].Name())

		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	sort.Sort(posts)

	return posts, err
}

func (f *FileRepository) FetchPost(filename string) (*BlogPost, error) {

	post := new(BlogPost)

	dirname := f.postDirectory + string(filepath.Separator)

	file, err := ioutil.ReadFile(dirname + filename)

	if err != nil {
		return post, err
	}

	file = []byte(f.extractHeader(string(file), post))

	htmlFlags := blackfriday.HTML_USE_SMARTYPANTS
	extensions := blackfriday.EXTENSION_HARD_LINE_BREAK | blackfriday.EXTENSION_FENCED_CODE | blackfriday.EXTENSION_NO_INTRA_EMPHASIS

	renderer := blackfriday.HtmlRenderer(htmlFlags, post.Title(), "")

	output := blackfriday.Markdown(file, renderer, extensions)

	post.SetBody(string(output))

	return post, nil
}

func (f* FileRepository) extractHeader(text string, post *BlogPost) string {

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
