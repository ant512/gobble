package main

import (
	"io/ioutil"
	"path/filepath"
)

type Comments []*Comment

func LoadComments(path string) (Comments, error) {
	files, err := ioutil.ReadDir(path)

	if err != nil {
		return nil, err
	}

	comments := Comments{}

	for _, file := range files {

		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) != ".md" {
			continue
		}

		comment, err := LoadComment(filepath.Join(path, file.Name()))

		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

func (c Comments) Len() int {
	return len(c)
}

func (c Comments) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c Comments) Less(i, j int) bool {
	return c[i].Metadata.Date.Before(c[j].Metadata.Date)
}

func (c Comments) ContainsTerm(term string) bool {
	for _, comment := range c {
		if comment.ContainsTerm(term) {
			return true
		}
	}

	return false
}
