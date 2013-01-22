package main

import (
	"github.com/russross/blackfriday"
	"io/ioutil"
	"strings"
)

type FileRepository struct {
	
}

func (f* FileRepository) FetchPost(id string) (*BlogPost, error) {

	post := new(BlogPost)

	file, err := ioutil.ReadFile(id + ".md")

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
			data := strings.Trim(components[1], " ")

			switch header {
				case "title":
					post.SetTitle(data)
				case "tags":
					post.SetTags(strings.Split(data, ","))
				case "date":
					post.SetPublishDate(data)
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
