package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type BlogPost struct {
	Title            string
	Id               int
	Path             string
	PublishDate      time.Time
	Tags             []string
	Body             string
	Comments         Comments
	DisallowComments bool
	Url              string
}

func LoadPost(path string) (*BlogPost, error) {

	b := &BlogPost{}
	b.Path = path

	file, err := ioutil.ReadFile(path)

	if err != nil {
		return b, err
	}

	file = []byte(strings.Replace(string(file), "\r", "", -1))
	file = []byte(b.extractHeader(string(file)))

	b.Body = convertMarkdownToHtml(&file)
	b.Url = b.urlFromTitle(b.Title)

	b.loadComments()

	return b, nil
}

func (b *BlogPost) NonSpamComments() Comments {
	comments := Comments{}

	for _, comment := range b.Comments {
		if !comment.IsSpam {
			comments = append(comments, comment)
		}
	}

	return comments
}

func (b *BlogPost) ContainsTag(tag string) bool {
	for _, t := range b.Tags {
		if t == strings.ToLower(tag) {
			return true
		}
	}

	return false
}

func (b *BlogPost) ContainsTerm(term string) bool {

	term = strings.ToLower(term)

	if b.ContainsTag(term) {
		return true
	}

	if b.Comments.ContainsTerm(term) {
		return true
	}

	terms := strings.Split(term, " ")
	body := strings.ToLower(b.Body)
	title := strings.ToLower(b.Title)

	for _, item := range terms {
		if !strings.Contains(body, item) && !strings.Contains(title, item) {
			return false
		}
	}

	return true
}

func (b *BlogPost) AllowsComments() bool {
	if b.DisallowComments {
		return false
	}

	if SharedConfig.CommentsOpenForDays == 0 {
		return true
	}

	var closeDate = b.PublishDate.Add(time.Hour * 24 * time.Duration(SharedConfig.CommentsOpenForDays))

	return time.Now().Before(closeDate)
}

func (b *BlogPost) extractHeader(text string) string {

	headerSize := parseHeader(text, func(key, value string) {
		switch key {
		case "title":
			b.Title = value
		case "id":
			b.Id, _ = strconv.Atoi(value)
		case "tags":

			tags := strings.Split(value, ",")

			formattedTags := []string{}

			for j := range tags {
				tags[j] = strings.Trim(tags[j], " ")
				tags[j] = strings.Replace(tags[j], " ", "-", -1)
				tags[j] = strings.Replace(tags[j], "/", "-", -1)
				tags[j] = strings.Replace(tags[j], "#", "", -1)
				tags[j] = strings.ToLower(tags[j])

				if tags[j] != "" {
					formattedTags = append(formattedTags, tags[j])
				}
			}

			b.Tags = formattedTags
		case "date":
			b.PublishDate = stringToTime(value)
		case "disallowcomments":
			b.DisallowComments = value == "true"
		default:
		}
	})

	return text[headerSize:]
}

func (b *BlogPost) urlFromTitle(title string) string {
	title = strings.ToLower(title)
	title = strings.Replace(title, " ", "-", -1)
	title = strings.Replace(title, ",", "", -1)
	title = strings.Replace(title, "#", "", -1)
	title = strings.Replace(title, ":", "", -1)
	title = strings.Replace(title, "\"", "", -1)
	title = strings.Replace(title, "?", "", -1)
	title = strings.Replace(title, "/", "", -1)

	return fmt.Sprintf("%04d/%02d/%02d/%s", b.PublishDate.Year(), b.PublishDate.Month(), b.PublishDate.Day(), title)
}

func (b *BlogPost) loadComments() {

	dirname := b.Path[:len(b.Path)-3] + string(filepath.Separator) + "comments" + string(filepath.Separator)

	files, err := ioutil.ReadDir(dirname)

	if err != nil {
		return
	}

	b.Comments = Comments{}

	for i := range files {

		if files[i].IsDir() {
			continue
		}

		if filepath.Ext(files[i].Name()) != ".md" {
			continue
		}

		comment, err := LoadComment(dirname + files[i].Name())

		if err != nil {
			log.Fatal(err)
			return
		}

		b.Comments = append(b.Comments, comment)
	}
}
