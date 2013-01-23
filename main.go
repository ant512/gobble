package main

import (
	"github.com/bmizerany/pat"
	"text/template"
	"net/http"
	"log"
)

func home(w http.ResponseWriter, req *http.Request) {

	repo := FileRepository{}
	repo.SetPostDirectory("./posts")

	posts, err := repo.FetchPostsInRange(0, 10)

	if err != nil {
		log.Println("Could not load post")
		return
	}

	t, _ := template.ParseFiles("./templates/home.html")
	t.Execute(w, posts)
}

func taggedPosts(w http.ResponseWriter, req *http.Request) {

	tag := req.URL.Query().Get(":tag")

	repo := FileRepository{}
	repo.SetPostDirectory("./posts")

	posts, err := repo.FetchPostsWithTag(tag)

	if err != nil {
		log.Println("Could not load posts")
		return
	}

	t, _ := template.ParseFiles("./templates/home.html")
	t.Execute(w, posts)
}

func tags(w http.ResponseWriter, req *http.Request) {
	repo := FileRepository{}
	repo.SetPostDirectory("./posts")

	tags, err := repo.FetchAllTags()

	if err != nil {
		log.Println("Could not load tags")
		return
	}

	t, _ := template.ParseFiles("./templates/tags.html")
	t.Execute(w, tags)
}

func main() {

	m := pat.New()
	m.Get("/tags/:tag", http.HandlerFunc(taggedPosts))
	m.Get("/tags/", http.HandlerFunc(tags))
	m.Get("/", http.HandlerFunc(home))

	http.Handle("/", m)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("img"))))
	http.ListenAndServe(":8080", nil)
}
