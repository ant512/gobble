package main

import (
	"github.com/russross/blackfriday"
	"io/ioutil"
	"os"
	"strings"
)

type metadataParseHandler func(key, value string)
type bodyParseHandler func(value string)
type fileInfoParseHandler func(fileInfo os.FileInfo)

func loadBlogFile(path string, fileInfoHandler fileInfoParseHandler, metadataHandler metadataParseHandler, bodyHandler bodyParseHandler) error {

	info, err := os.Stat(path)

	if err != nil {
		return err
	}

	fileInfoHandler(info)

	file, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	parseBlogFile(string(file), metadataHandler, bodyHandler)

	return nil
}

func parseBlogFile(text string, metadataHandler metadataParseHandler, bodyHandler bodyParseHandler) {
	text = strings.Replace(text, "\r", "", -1)

	headerSize := parseBlogFileHeader(text, metadataHandler)
	body := text[headerSize:]
	bodyHandler(body)
}

func parseBlogFileHeader(text string, handler metadataParseHandler) int {

	lines := strings.Split(text, "\n")
	headerSize := 0

	for _, line := range lines {
		if strings.Contains(line, ":") {
			components := strings.Split(line, ":")
			key := strings.ToLower(strings.TrimSpace(components[0]))
			separatorIndex := strings.Index(line, ":") + 1
			value := strings.TrimSpace(line[separatorIndex:])

			headerSize += len(line) + 1

			handler(key, value)
		} else {
			break
		}
	}

	return headerSize
}

func convertMarkdownToHtml(markdown *[]byte) string {
	htmlFlags := blackfriday.HTML_USE_SMARTYPANTS
	extensions := blackfriday.EXTENSION_AUTOLINK | blackfriday.EXTENSION_FENCED_CODE | blackfriday.EXTENSION_NO_INTRA_EMPHASIS | blackfriday.EXTENSION_STRIKETHROUGH

	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	output := blackfriday.Markdown(*markdown, renderer, extensions)

	return string(output)
}
