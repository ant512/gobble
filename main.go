package main

import (
	"github.com/bmizerany/pat"
	"text/template"
	"net/http"
	"log"
	"fmt"
)

func home(w http.ResponseWriter, req *http.Request) {

	repo := FileRepository{}
	repo.SetPostDirectory("./posts")

	post, err := repo.FetchNewestPost()

	if err != nil {
		log.Println("Could not load post")
		return
	}

	t, _ := template.ParseFiles("./templates/home.html")
	t.Execute(w, post)
}

func tags(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, req.URL.Query().Get(":tag"))
}

func main() {

	m := pat.New()
	m.Get("/tags/:tag", http.HandlerFunc(tags))
	m.Get("/", http.HandlerFunc(home))

	http.Handle("/", m)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("img"))))
	http.ListenAndServe(":8080", nil)
}
