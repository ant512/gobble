package main

import (
	"fmt"
	"github.com/bmizerany/pat"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"flag"
	"path/filepath"
	"os"
)

const postsPerPage = 10
const version = "1.0"

var repo *FileRepository
var config *Config
var themePath string

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
			nextURL = fmt.Sprintf("/?search=%v&page=%v", term, pageNumber)
		} else {
			nextURL = fmt.Sprintf("/?page=%v", pageNumber)
		}
	}

	if int(pageNumber) < count / postsPerPage {
		if len(term) > 0 {
			previousURL = fmt.Sprintf("/?search=%v&page=%v", term, pageNumber + 2)
		} else {
			previousURL = fmt.Sprintf("/?page=%v", pageNumber + 2)
		}
	}

	page := struct {
		Posts BlogPosts
		NextURL string
		PreviousURL string
		SiteName string
	} {
		posts,
		nextURL,
		previousURL,
		config.Name,
	}

	t, _ := template.ParseFiles(themePath + "/templates/home.html")
	t.Execute(w, page)
}

func taggedPosts(w http.ResponseWriter, req *http.Request) {

	tag := req.URL.Query().Get(":tag")
	pageNumber, err := strconv.ParseInt(req.URL.Query().Get(":page"), 10, 32)

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

	posts, count := repo.PostsWithTag(tag, int(pageNumber) * postsPerPage, postsPerPage)

	if pageNumber > 0 {
		nextURL = fmt.Sprintf("/tags/%v/%v", tag, pageNumber)
	}

	if int(pageNumber) < count / postsPerPage {
		previousURL = fmt.Sprintf("/tags/%v/%v", tag, pageNumber + 2)
	}

	page := struct {
		Posts BlogPosts
		NextURL string
		PreviousURL string
		SiteName string
	} {
		posts,
		nextURL,
		previousURL,
		config.Name,
	}

	t, _ := template.ParseFiles(themePath + "/templates/home.html")
	t.Execute(w, page)
}

func tags(w http.ResponseWriter, req *http.Request) {

	tags := repo.AllTags()

	page := struct {
		Tags []string
		SiteName string
	} {
		tags,
		config.Name,
	}

	t, _ := template.ParseFiles(themePath + "/templates/tags.html")
	t.Execute(w, page)
}

func archive(w http.ResponseWriter, req *http.Request) {

	posts := repo.AllPosts()

	page := struct {
		Posts BlogPosts
		SiteName string
	} {
		posts,
		config.Name,
	}

	t, _ := template.ParseFiles(themePath + "/templates/archive.html")
	t.Execute(w, page)
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

	page := struct {
		Post *BlogPost
		SiteName string
	} {
		post,
		config.Name,
	}

	t, _ := template.ParseFiles(themePath + "/templates/post.html")
	t.Execute(w, page)
}

func rss(w http.ResponseWriter, req *http.Request) {

	posts, _ := repo.SearchPosts("", 0, 10)

	page := struct {
		Posts BlogPosts
		SiteName string
		SiteDescription string
		SiteAddress string
	} {
		posts,
		config.Name,
		config.Description,
		config.Address,
	}

	t, _ := template.ParseFiles(themePath + "/templates/rss.html")
	t.Execute(w, page)
}

func printInfo() {
	fmt.Printf("Gobble Blogging Engine (version %v)\n", version)
	fmt.Println("http://simianzombie.com")
	fmt.Println("")
	fmt.Println("Copyright (C) 2013 Antony Dzeryn")
	fmt.Println("")
}

func loadConfig() {
	configPath := flag.String("config", "./gobble.conf", "config file path")
	flag.Parse()

	var err error

	config, err = LoadConfig(*configPath)

	if err != nil {
		log.Println("Could not load config file", *configPath)
		log.Fatal(err)
	}

	themePath = "themes" + string(filepath.Separator) + config.Theme

	_, err = os.Stat(themePath)

	if err != nil {
		log.Fatal("Could not load theme", themePath)
	}

	_, err = os.Stat(config.PostPath)

	if err != nil {
		log.Fatal("Could not load posts from", config.PostPath)
	}
}

func prepareHandler() {

	repo = NewFileRepository(config.PostPath)

	m := pat.New()
	m.Get("/tags/:tag/:page", http.HandlerFunc(taggedPosts))
	m.Get("/tags/:tag", http.HandlerFunc(taggedPosts))
	m.Get("/tags/", http.HandlerFunc(tags))
	m.Get("/archive/", http.HandlerFunc(archive))
	m.Get("/rss/", http.HandlerFunc(rss))
	m.Get("/posts/:year/:month/:day/:title", http.HandlerFunc(post))
	m.Get("/", http.HandlerFunc(home))

	http.Handle("/", m)
	http.Handle("/theme/", http.StripPrefix("/theme/", http.FileServer(http.Dir(themePath))))
	http.Handle("/rainbow/", http.StripPrefix("/rainbow/", http.FileServer(http.Dir("rainbow"))))

	fmt.Printf("Listening on port %v\n", config.Port)
	fmt.Printf("Using theme \"%v\"\n", config.Theme)
	fmt.Printf("Post data stored in \"%v\"\n", config.PostPath)

	http.ListenAndServe(":" + strconv.FormatInt(config.Port, 10), nil)
}

func main() {
	printInfo()
	loadConfig()

	valid, err := ValidateComment("This is some text", config.Address, "127.0.0.1", "curl", "viagra-test-123", "joe@example.com", config.AkismetAPIKey)

	if err != nil {
		log.Println(err)
	}

	if valid {
		log.Println("Comment is valid")
	} else {
		log.Println("Comment is invalid")
	}

	prepareHandler()
}
