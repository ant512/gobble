package main

import (
	"fmt"
	"github.com/ant512/gobble/akismet"
	"html"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BlogPost struct {
	Title            string
	Id               int
	PostPath         string
	CommentPath      string
	PublishDate      time.Time
	Tags             []string
	Body             string
	RawBody          string
	Comments         Comments
	DisallowComments bool
	Url              string
	Filename         string
	mutex            sync.RWMutex
}

func LoadPost(filename, postPath, commentPath string) (*BlogPost, error) {

	b := &BlogPost{}
	b.PostPath = postPath
	b.Filename = filename
	b.CommentPath = commentPath

	fullPath := filepath.Join(postPath, filename)

	err := loadBlogFile(fullPath, func(key, value string) {
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
	}, func(value string) {
		bytes := []byte(value)

		b.RawBody = value
		b.Body = convertMarkdownToHtml(&bytes)
	})

	if (err == nil) {
		b.Url = b.urlFromBlogPostProperties()
		b.loadComments()
	} else {
		log.Println(err)
	}

	return b, err
}

func (b *BlogPost) NonSpamComments() Comments {
	comments := Comments{}

	for _, comment := range b.Comments {
		if !comment.Metadata.IsSpam {
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
	body := strings.ToLower(b.RawBody)
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

func (b *BlogPost) SaveComment(akismetAPIKey, serverAddress, remoteAddress, userAgent, referrer, author, email, body string) {

	// TODO: Ensure file name is unique
	isSpam, _ := akismet.IsSpamComment(body, serverAddress, remoteAddress, userAgent, referrer, author, email, akismetAPIKey)
	comment := NewComment(html.EscapeString(author), html.EscapeString(email), html.EscapeString(body), isSpam)

	b.mutex.Lock()
	b.Comments = append(b.Comments, comment)
	b.mutex.Unlock()

	commentPath := filepath.Join(b.CommentPath, b.Filename[:len(b.Filename)-3])
	filename := timeToFilename(comment.Metadata.Date)
	fullPath := filepath.Join(commentPath, filename)

	log.Println(commentPath)
	os.MkdirAll(commentPath, 0775)

	content := comment.String()

	err := ioutil.WriteFile(fullPath, []byte(content), 0644)

	if err != nil {
		log.Println(err)
	}
}

func (b *BlogPost) urlFromBlogPostProperties() string {
	title := strings.ToLower(b.Title)
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

	filename := b.Filename[:len(b.Filename)-3]
	dirname := b.CommentPath + string(filepath.Separator) + filename + string(filepath.Separator)

	b.Comments, _ = LoadComments(dirname)
}
