package main

import (
	"fmt"
	"github.com/bmizerany/pat"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

func home(w http.ResponseWriter, req *http.Request) {

	posts := repo.PostsInRange(0, 10)

	t, _ := template.ParseFiles("./templates/home.html")
	t.Execute(w, posts)
}

func taggedPosts(w http.ResponseWriter, req *http.Request) {

	tag := req.URL.Query().Get(":tag")

	posts := repo.PostsWithTag(tag)

	t, _ := template.ParseFiles("./templates/home.html")
	t.Execute(w, posts)
}

func tags(w http.ResponseWriter, req *http.Request) {

	tags := repo.AllTags()

	t, _ := template.ParseFiles("./templates/tags.html")
	t.Execute(w, tags)
}

func archive(w http.ResponseWriter, req *http.Request) {

	posts := repo.AllPosts()

	t, _ := template.ParseFiles("./templates/archive.html")
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

	t, _ := template.ParseFiles("./templates/post.html")
	t.Execute(w, post)
}

func rss(w http.ResponseWriter, req *http.Request) {

	posts := repo.PostsInRange(0, 10)

	t, _ := template.ParseFiles("./templates/rss.html")
	t.Execute(w, posts)
}

var repo *FileRepository

func main() {

	repo = NewFileRepository("./posts")

	m := pat.New()
	m.Get("/tags/:tag", http.HandlerFunc(taggedPosts))
	m.Get("/tags/", http.HandlerFunc(tags))
	m.Get("/archive/", http.HandlerFunc(archive))
	m.Get("/rss/", http.HandlerFunc(rss))
	m.Get("/posts/:year/:month/:day/:title", http.HandlerFunc(post))
	m.Get("/", http.HandlerFunc(home))

	http.Handle("/", m)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("img"))))
	http.Handle("/rainbow/", http.StripPrefix("/rainbow/", http.FileServer(http.Dir("rainbow"))))

	http.ListenAndServe(":8080", nil)
}
