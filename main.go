package main

import (
	"log"
)

func main() {

	repo := FileRepository{}
	repo.SetPostDirectory("./posts")

	posts, err := repo.FetchAllPosts()

	if err != nil {
		log.Println("Could not load posts")
		return
	}

	for i := range posts {

		post := posts[i]

		log.Println(post.Title())
		log.Println(post.PublishDate())
		log.Println(post.Tags())
		log.Println(post.Body())
	}
}

