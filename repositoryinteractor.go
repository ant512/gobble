package main

type RepositoryInteractor interface {
	FetchPost(id string) (BlogPost, error)
}
