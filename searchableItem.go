package main

type SearchableItem interface {
	ContainsTerm(term string) bool
}
