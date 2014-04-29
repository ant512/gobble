package main

import (
	"github.com/ant512/gobble/akismet"
	"github.com/russross/blackfriday"
	"html"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type FileRepository struct {
	directory string
	posts     BlogPosts
	mutex     sync.RWMutex
}

func NewFileRepository(directory string) *FileRepository {

	f := new(FileRepository)
	f.directory = directory

	return f
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

func convertMarkdownToHtml(markdown *[]byte) string {
	htmlFlags := blackfriday.HTML_USE_SMARTYPANTS
	extensions := blackfriday.EXTENSION_AUTOLINK | blackfriday.EXTENSION_HARD_LINE_BREAK | blackfriday.EXTENSION_FENCED_CODE | blackfriday.EXTENSION_NO_INTRA_EMPHASIS

	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	output := blackfriday.Markdown(*markdown, renderer, extensions)

	return string(output)
}
