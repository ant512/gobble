package main

import (
	"text/template"
	"net/http"
	"log"
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

func main() {
	http.HandleFunc("/", home)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("img"))))
	http.ListenAndServe(":8080", nil)
}
