package main

import (
	"fmt"
	"github.com/bmizerany/pat"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"flag"
)

const postsPerPage = 10

var repo *FileRepository
var port *int64

func home(w http.ResponseWriter, req *http.Request) {

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

	posts, count := repo.SearchPosts(term, int(pageNumber) * postsPerPage, postsPerPage)

	if pageNumber > 0 {
		if len(term) > 0 {
			previousURL = fmt.Sprintf("/?search=%v&page=%v", term, pageNumber)
		} else {
			previousURL = fmt.Sprintf("/?page=%v", pageNumber)
		}
	}

	if int(pageNumber) < count / postsPerPage {
		if len(term) > 0 {
			nextURL = fmt.Sprintf("/?search=%v&page=%v", term, pageNumber + 2)
		} else {
			nextURL = fmt.Sprintf("/?page=%v", pageNumber + 2)
		}
	}

	page := struct {
		Posts BlogPosts
		NextURL string
		PreviousURL string
	} {
		posts,
		nextURL,
		previousURL,
	}

	t, _ := template.ParseFiles("./theme/templates/home.html")
	t.Execute(w, page)
}

func taggedPosts(w http.ResponseWriter, req *http.Request) {

	tag := req.URL.Query().Get(":tag")

	posts := repo.PostsWithTag(tag)

	t, _ := template.ParseFiles("./theme/templates/home.html")
	t.Execute(w, posts)
}

func tags(w http.ResponseWriter, req *http.Request) {

	tags := repo.AllTags()

	t, _ := template.ParseFiles("./theme/templates/tags.html")
	t.Execute(w, tags)
}

func archive(w http.ResponseWriter, req *http.Request) {

	posts := repo.AllPosts()

	t, _ := template.ParseFiles("./theme/templates/archive.html")
	t.Execute(w, posts)
}

func post(w http.ResponseWriter, req *http.Request) {

	title := req.URL.Query().Get(":title")

	year, err := strconv.Atoi(req.URL.Query().Get(":year"))

	if err != nil {
		log.Println("Invalid year supplied")
		return
	}

	month, err := strconv.Atoi(req.URL.Query().Get(":month"))

	if err != nil {
		log.Println("Invalid month supplied")
		return
	}

	day, err := strconv.Atoi(req.URL.Query().Get(":day"))

	if err != nil {
		log.Println("Invalid day supplied")
		return
	}

	url := fmt.Sprintf("%04d/%02d/%02d/%s", year, month, day, title)

	post, err := repo.PostWithUrl(url)

	if err != nil {
		log.Println("Could not load post")
		return
	}

	t, _ := template.ParseFiles("./theme/templates/post.html")
	t.Execute(w, post)
}

func rss(w http.ResponseWriter, req *http.Request) {

	posts, _ := repo.SearchPosts("", 0, 10)

	t, _ := template.ParseFiles("./theme/templates/rss.html")
	t.Execute(w, posts)
}

func main() {

	repo = NewFileRepository("./posts")

	port = flag.Int64("port", 8080, "port number")
	flag.Parse()

	var version = "1.0"

	m := pat.New()
	m.Get("/tags/:tag", http.HandlerFunc(taggedPosts))
	m.Get("/tags/", http.HandlerFunc(tags))
	m.Get("/archive/", http.HandlerFunc(archive))
	m.Get("/rss/", http.HandlerFunc(rss))
	m.Get("/posts/:year/:month/:day/:title", http.HandlerFunc(post))
	m.Get("/", http.HandlerFunc(home))

	http.Handle("/", m)
	http.Handle("/theme/", http.StripPrefix("/theme/", http.FileServer(http.Dir("theme"))))
	http.Handle("/rainbow/", http.StripPrefix("/rainbow/", http.FileServer(http.Dir("rainbow"))))

	fmt.Printf("Gobble Blogging Engine (version %v)\n", version)
	fmt.Println("http://simianzombie.com")
	fmt.Println("")
	fmt.Println("Copyright (C) 2013 Antony Dzeryn")
	fmt.Println("")
	fmt.Printf("Listening on port %v\n", *port)

	http.ListenAndServe(":" + strconv.FormatInt(*port, 10), nil)
}
