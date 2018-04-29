package main

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"text/template"
	"time"
)

func pageNumberFromRequest(req *http.Request, query string) int64 {
	pageNumber, err := strconv.ParseInt(req.URL.Query().Get(query), 10, 32)

	if err != nil {
		pageNumber = 0
	} else {

		// We want to use 0-based page numbers internally, but expose them as
		// 1-based.
		pageNumber -= 1

		if pageNumber < 0 {
			pageNumber = 0
		}
	}

	return pageNumber
}

func home(w http.ResponseWriter, req *http.Request) {

	id, err := strconv.Atoi(req.URL.Query().Get("p"))

	if err == nil {

		post, _ := blog.PostWithId(id)
		showSinglePost(post, w, req)

		return
	}

	term := req.URL.Query().Get("search")
	pageNumber := pageNumberFromRequest(req, "page")

	var previousURL string
	var nextURL string

	posts, count := blog.SearchPosts(term, int(pageNumber)*SharedConfig.PostsPerPage, SharedConfig.PostsPerPage)

	if pageNumber > 0 {
		if len(term) > 0 {
			nextURL = fmt.Sprintf("/?search=%v&page=%v", term, pageNumber)
		} else {
			nextURL = fmt.Sprintf("/?page=%v", pageNumber)
		}
	}

	if float64(pageNumber+1) < float64(count)/float64(SharedConfig.PostsPerPage) {
		if len(term) > 0 {
			previousURL = fmt.Sprintf("/?search=%v&page=%v", term, pageNumber+2)
		} else {
			previousURL = fmt.Sprintf("/?page=%v", pageNumber+2)
		}
	}

	var searchPlaceholder string

	if len(term) > 0 {
		searchPlaceholder = html.EscapeString(term)
	} else {
		searchPlaceholder = "Search"
	}

	page := struct {
		Posts             BlogPosts
		Config            *Config
		NextURL           string
		PreviousURL       string
		SearchPlaceholder string
	}{
		posts,
		SharedConfig,
		nextURL,
		previousURL,
		searchPlaceholder,
	}

	t, _ := template.ParseFiles(SharedConfig.FullThemePath() + "/templates/home.html")
	t.Execute(w, page)
}

func taggedPosts(w http.ResponseWriter, req *http.Request) {

	tag := req.URL.Query().Get(":tag")
	pageNumber := pageNumberFromRequest(req, ":page")

	var previousURL string
	var nextURL string

	posts, count := blog.PostsWithTag(tag, int(pageNumber)*SharedConfig.PostsPerPage, SharedConfig.PostsPerPage)

	if pageNumber > 0 {
		nextURL = fmt.Sprintf("/tags/%v/%v", tag, pageNumber)
	}

	if float64(pageNumber+1) < float64(count)/float64(SharedConfig.PostsPerPage) {
		previousURL = fmt.Sprintf("/tags/%v/%v", tag, pageNumber+2)
	}

	page := struct {
		Posts             BlogPosts
		Config            *Config
		NextURL           string
		PreviousURL       string
		SearchPlaceholder string
	}{
		posts,
		SharedConfig,
		nextURL,
		previousURL,
		"",
	}

	t, _ := template.ParseFiles(SharedConfig.FullThemePath() + "/templates/home.html")
	t.Execute(w, page)
}

func tags(w http.ResponseWriter, req *http.Request) {

	tags := blog.AllTags()

	page := struct {
		Tags   map[string]int
		Config *Config
	}{
		tags,
		SharedConfig,
	}

	t, _ := template.ParseFiles(SharedConfig.FullThemePath() + "/templates/tags.html")
	t.Execute(w, page)
}

func archive(w http.ResponseWriter, req *http.Request) {

	posts := blog.AllPosts()

	page := struct {
		Posts  BlogPosts
		Config *Config
	}{
		posts,
		SharedConfig,
	}

	t, _ := template.ParseFiles(SharedConfig.FullThemePath() + "/templates/archive.html")
	t.Execute(w, page)
}

func rss(w http.ResponseWriter, req *http.Request) {

	posts, _ := blog.SearchPosts("", 0, 10)
	var updated time.Time

	if len(posts) > 0 {
		updated = posts[0].Metadata.Date
	}

	page := struct {
		Posts   BlogPosts
		Updated time.Time
		Config  *Config
	}{
		posts,
		updated,
		SharedConfig,
	}

	t, _ := template.ParseFiles(SharedConfig.FullThemePath() + "/templates/rss.html")
	t.Execute(w, page)
}
