package main

import (
	"log"
)

func main() {

	repo := FileRepository{}

	post, err := repo.FetchPost("1")

	if err != nil {
		log.Println("Could not load post")
		return
	}

	log.Println(post.Title())
	log.Println(post.PublishDate())
	log.Println(post.Tags())
	log.Println(post.Body())
}

