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

func home(w http.ResponseWriter, req *http.Request) {

	term := req.URL.Query().Get("search")
	page, err := strconv.ParseInt(req.URL.Query().Get("page"), 10, 32)

	if err != nil {
		page = 0
	}

	log.Println(term)

	posts := repo.SearchPosts(term, int(page) * 10, 10)

	t, _ := template.ParseFiles("./theme/templates/home.html")
	t.Execute(w, posts)
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

	posts := repo.SearchPosts("", 0, 10)

	t, _ := template.ParseFiles("./theme/templates/rss.html")
	t.Execute(w, posts)
}

var repo *FileRepository
var port *int64

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
