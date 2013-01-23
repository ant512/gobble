package main

import (
	"github.com/russross/blackfriday"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
	"strconv"
)

type FileRepository struct {
	postDirectory string
}

func (f *FileRepository) PostDirectory() string {
	return f.postDirectory
}

func (f *FileRepository) SetPostDirectory(s string) {
	f.postDirectory = s
}

func (f *FileRepository) FetchNewestPost() (*BlogPost, error) {
	dirname := f.postDirectory + string(filepath.Separator)

	files, err := ioutil.ReadDir(dirname)

	var newestPost *BlogPost = nil

	for i := range files {

		post, err := f.FetchPost(files[i].Name())

		if err != nil {
			return nil, err
		}

		if newestPost == nil || post.PublishDate().After(newestPost.PublishDate()) {
			newestPost = post
		}
	}

	return newestPost, err
}

func (f *FileRepository) FetchAllPosts() ([]*BlogPost, error) {

	dirname := f.postDirectory + string(filepath.Separator)

	files, err := ioutil.ReadDir(dirname)

	posts := []*BlogPost{}

	for i := range files {

		post, err := f.FetchPost(files[i].Name())

		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}



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
	extensions := 0

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
					post.SetTags(strings.Split(data, ","))
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
