package main

type Comments []*Comment

func (c Comments) Len() int {
	return len(c)
}

func (c Comments) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c Comments) Less(i, j int) bool {
	return c[i].Date.Before(c[j].Date)
}
