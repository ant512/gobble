package main

import (
	"flag"
	"fmt"
	"github.com/bmizerany/pat"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

const postsPerPage = 10
const version = "2.0"

var blog *Blog
var repo *FileRepository
var SharedConfig *Config

func printInfo() {
	fmt.Printf("Gobble Blogging Engine (version %v)\n", version)
	fmt.Println("http://simianzombie.com")
	fmt.Println("")
	fmt.Println("Copyright (C) 2013-2014 Antony Dzeryn")
	fmt.Println("")
}

func prepareHandler() {

	m := pat.New()
	m.Get("/tags/:tag/:page", http.HandlerFunc(taggedPosts))
	m.Get("/tags/:tag", http.HandlerFunc(taggedPosts))
	m.Get("/tags/", http.HandlerFunc(tags))
	m.Get("/archive/", http.HandlerFunc(archive))
	m.Get("/rss", http.HandlerFunc(rss))
	m.Get("/posts/:year/:month/:day/:title", http.HandlerFunc(post))

	for key, value := range SharedConfig.StaticFiles {
		func(url, path string) {
			m.Get(url, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				log.Println("Static file request: " + url + " Serving: " + path)
				http.ServeFile(w, req, SharedConfig.StaticFilePath+string(filepath.Separator)+path)
			}))
		}(key, value)
	}

	m.Get("/", http.HandlerFunc(home))

	m.Post("/posts/:year/:month/:day/:title/comments", http.HandlerFunc(createComment))

	http.Handle("/", m)
	http.Handle("/theme/", http.StripPrefix("/theme/", http.FileServer(http.Dir(SharedConfig.FullThemePath()))))
	http.Handle("/highlight/", http.StripPrefix("/highlight/", http.FileServer(http.Dir("highlight"))))
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(SharedConfig.MediaPath))))

	fmt.Printf("Listening on port %v\n", SharedConfig.Port)
	fmt.Printf("Using theme \"%v\"\n", SharedConfig.Theme)
	fmt.Printf("Post data stored in \"%v\"\n", SharedConfig.PostPath)
	fmt.Printf("Media stored in \"%v\"\n", SharedConfig.MediaPath)
	fmt.Printf("Themes stored in \"%v\"\n", SharedConfig.ThemePath)
	fmt.Printf("Files stored in \"%v\"\n", SharedConfig.StaticFilePath)

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
