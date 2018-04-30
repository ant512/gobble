package main

import (
	"fmt"
	"github.com/dpapathanasiou/go-recaptcha"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
)

const maxCommentNameLength = 254
const maxCommentEmailLength = 254
const maxCommentBodyLength = 5000

func showSinglePost(b *BlogPost, w http.ResponseWriter, req *http.Request) {

	if b == nil {
		http.NotFound(w, req)
		return
	}

	page := PostPage{}
	page.Post = b
	page.Config = SharedConfig
	page.CommentName = ""
	page.CommentEmail = ""
	page.CommentBody = ""
	page.CommentNameError = ""
	page.CommentEmailError = ""
	page.CommentBodyError = ""

	t, _ := template.ParseFiles(SharedConfig.FullThemePath() + "/templates/post.html")
	t.Execute(w, page)
}

func postWithQuery(query url.Values) (*BlogPost, error) {

	title := query.Get(":title")

	year, err := strconv.Atoi(query.Get(":year"))

	if err != nil {
		log.Println("Invalid year supplied")
		return nil, err
	}

	month, err := strconv.Atoi(query.Get(":month"))

	if err != nil {
		log.Println("Invalid month supplied")
		return nil, err
	}

	day, err := strconv.Atoi(query.Get(":day"))

	if err != nil {
		log.Println("Invalid day supplied")
		return nil, err
	}

	url := fmt.Sprintf("%04d/%02d/%02d/%s", year, month, day, title)

	post, err := blog.PostWithUrl(url)

	return post, err
}

func post(w http.ResponseWriter, req *http.Request) {
	post, _ := postWithQuery(req.URL.Query())
	showSinglePost(post, w, req)
}

func createComment(w http.ResponseWriter, req *http.Request) {

	post, err := postWithQuery(req.URL.Query())

	if err != nil {
		log.Println("Could not load post")
		return
	}

	if post.Metadata.DisallowComments {
		return
	}

	author := strings.TrimSpace(req.FormValue("name"))
	email := strings.TrimSpace(req.FormValue("email"))
	body := strings.TrimSpace(req.FormValue("comment"))

	hasErrors := false
	commentNameError := ""
	commentEmailError := ""
	commentBodyError := ""
	commentRecaptchaError := ""

	if len(author) == 0 {
		hasErrors = true
		commentNameError = "Name cannot be blank"
	} else if len(author) > 254 {
		hasErrors = true
		commentNameError = fmt.Sprintf("Name must be less than %v characters", +maxCommentNameLength)
	}

	if len(email) < 5 {
		hasErrors = true
		commentEmailError = "Email must be a valid address"
	} else if len(email) > maxCommentEmailLength {
		hasErrors = true
		commentEmailError = fmt.Sprintf("Email must be less than %v characters", maxCommentEmailLength)
	} else if !strings.Contains(email, "@") {

		// Since regex is useless for validating emails, we'll just check for
		// the @ symbol.

		hasErrors = true
		commentEmailError = "Email must be a valid address"
	}

	if len(body) == 0 {
		hasErrors = true
		commentBodyError = "Comment cannot be blank"
	} else if len(body) > maxCommentBodyLength {
		hasErrors = true
		commentBodyError = fmt.Sprintf("Comment must be less than %v characters", maxCommentBodyLength)
	}

	if len(SharedConfig.RecaptchaPrivateKey) > 0 {
		recaptcha.Init(SharedConfig.RecaptchaPrivateKey)
		ip := strings.Split(req.RemoteAddr, ":")[0]

		success, _ := recaptcha.Confirm(ip, req.FormValue("g-recaptcha-response"))

		if (!success) {
			hasErrors = true
			commentRecaptchaError = "Incorrect reCAPTCHA entered"
		}
	}

	if !hasErrors {
		post.SaveComment(SharedConfig.AkismetAPIKey, SharedConfig.Address, getIpAddress(req), req.UserAgent(), req.Referer(), author, email, body)
		http.Redirect(w, req, "/posts/"+post.Url+"#comments", http.StatusFound)

		return
	} else {

		page := PostPage{}
		page.Post = post
		page.Config = SharedConfig
		page.CommentName = author
		page.CommentEmail = email
		page.CommentBody = body
		page.CommentNameError = commentNameError
		page.CommentEmailError = commentEmailError
		page.CommentBodyError = commentBodyError
		page.CommentRecaptchaError = commentRecaptchaError

		t, _ := template.ParseFiles(SharedConfig.FullThemePath() + "/templates/post.html")
		t.Execute(w, page)
	}
}

func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

func getIpAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIp := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIp == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		// TODO: should return first non-local address
		return parts[0]
	}
	return hdrRealIp
}
