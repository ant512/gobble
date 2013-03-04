package main

type PostPage struct {
	Post              *BlogPost
	Config            *Config
	CommentName       string
	CommentEmail      string
	CommentBody       string
	CommentNameError  string
	CommentEmailError string
	CommentBodyError  string
}
