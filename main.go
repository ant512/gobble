package main

import (
	"log"
)

func main() {

	repo := FileRepository{}
	repo.SetPostDirectory("./posts")

	post, err := repo.FetchNewestPost()

	if err != nil {
		log.Println("Could not load post")
		return
	}

	log.Println(post.Title())
	log.Println(post.PublishDate())
	log.Println(post.Tags())
	log.Println(post.Body())
}

