package main

import (
	"flag"
	"fmt"
	"github.com/bmizerany/pat"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"
)

const postsPerPage = 10
const version = "2.0"

var blog *Blog
var repo *FileRepository
var SharedConfig *Config

func home(w http.ResponseWriter, req *http.Request) {

	id, err := strconv.Atoi(req.URL.Query().Get("p"))

	if err == nil {

		post, _ := blog.PostWithId(id)
		showSinglePost(post, w, req)

		return
	}

	term := req.URL.Query().Get("search")
	pageNumber, err := strconv.ParseInt(req.URL.Query().Get("page"), 10, 32)

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

	var previousURL string
	var nextURL string

	posts, count := blog.SearchPosts(term, int(pageNumber)*postsPerPage, postsPerPage)

	if pageNumber > 0 {
		if len(term) > 0 {
			nextURL = fmt.Sprintf("/?search=%v&page=%v", term, pageNumber)
		} else {
			nextURL = fmt.Sprintf("/?page=%v", pageNumber)
		}
	}

	if int(pageNumber+1) < count/postsPerPage {
		if len(term) > 0 {
			previousURL = fmt.Sprintf("/?search=%v&page=%v", term, pageNumber+2)
		} else {
			previousURL = fmt.Sprintf("/?page=%v", pageNumber+2)
		}
	}

	var searchPlaceholder string

	if len(term) > 0 {
		searchPlaceholder = term
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
	pageNumber, err := strconv.ParseInt(req.URL.Query().Get(":page"), 10, 32)

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

	var previousURL string
	var nextURL string

	posts, count := blog.PostsWithTag(tag, int(pageNumber)*postsPerPage, postsPerPage)

	if pageNumber > 0 {
		nextURL = fmt.Sprintf("/tags/%v/%v", tag, pageNumber)
	}

	if int(pageNumber) < count/postsPerPage {
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
		updated = posts[0].PublishDate
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

func printInfo() {
	fmt.Printf("Gobble Blogging Engine (version %v)\n", version)
	fmt.Println("http://simianzombie.com")
	fmt.Println("")
	fmt.Println("Copyright (C) 2013-2014 Antony Dzeryn")
	fmt.Println("")
}

func favicon(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, SharedConfig.FullThemePath()+"/favicon.ico")
}

func robots(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, SharedConfig.FullThemePath()+"/robots.txt")
}

func prepareHandler() {

	m := pat.New()
	m.Get("/tags/:tag/:page", http.HandlerFunc(taggedPosts))
	m.Get("/tags/:tag", http.HandlerFunc(taggedPosts))
	m.Get("/tags/", http.HandlerFunc(tags))
	m.Get("/archive/", http.HandlerFunc(archive))
	m.Get("/rss", http.HandlerFunc(rss))
	m.Get("/posts/:year/:month/:day/:title", http.HandlerFunc(post))
	m.Get("/favicon.ico", http.HandlerFunc(favicon))
	m.Get("/robots.txt", http.HandlerFunc(robots))
	m.Get("/", http.HandlerFunc(home))

	m.Post("/posts/:year/:month/:day/:title/comments", http.HandlerFunc(createComment))

	http.Handle("/", m)
	http.Handle("/theme/", http.StripPrefix("/theme/", http.FileServer(http.Dir(SharedConfig.FullThemePath()))))
	http.Handle("/rainbow/", http.StripPrefix("/rainbow/", http.FileServer(http.Dir("rainbow"))))
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(SharedConfig.MediaPath))))

	fmt.Printf("Listening on port %v\n", SharedConfig.Port)
	fmt.Printf("Using theme \"%v\"\n", SharedConfig.Theme)
	fmt.Printf("Post data stored in \"%v\"\n", SharedConfig.PostPath)
	fmt.Printf("Media stored in \"%v\"\n", SharedConfig.MediaPath)
	fmt.Printf("Themes stored in \"%v\"\n", SharedConfig.ThemePath)

	postCount := len(blog.posts)

	if postCount == 1 {
		fmt.Printf("Serving 1 post")
	} else {
		fmt.Printf("Serving %d posts", postCount)
	}

	commentCount := 0

	for _, p := range blog.posts {
		commentCount += len(p.Comments)
	}

	if commentCount == 0 {
		fmt.Printf("\n")
	} else if commentCount == 1 {
		fmt.Printf(" and 1 comment\n")
	} else {
		fmt.Printf(" and %d comments\n", commentCount)
	}

	http.ListenAndServe(":"+strconv.FormatInt(SharedConfig.Port, 10), nil)
}

func main() {
	printInfo()

	configPath := flag.String("config", "./gobble.conf", "config file path")
	flag.Parse()

	var err error
	SharedConfig, err = LoadConfig(*configPath)

	if err != nil {
		log.Fatal(err)
	}

	blog, err = LoadBlog(SharedConfig.PostPath)

	if err != nil {
		log.Fatal(err)
	}

	repo = NewFileRepository(SharedConfig.PostPath)

	prepareHandler()
}
