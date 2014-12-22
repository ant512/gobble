package main

import (
	"log"
)

type Tags struct {
	tags map[string]int
}

func NewTags() Tags {
	t := Tags{}
	t.tags = make(map[string]int)

	return t
}

func (t Tags) AllTags() map[string]int {
	return t.tags
}

func (t Tags) AddTags(tags []string) {
	for _, tag := range tags {
		value := 1
		value += t.tags[tag]
		t.tags[tag] = value

		log.Println("Stored tag", tag, "with value", value)
	}
}

func (t Tags) RemoveTags(tags []string) {
	log.Println(t.tags)

	for _, tag := range tags {
		value := t.tags[tag] - 1

		if value == 0 {
			log.Println("Tag no longer used:", tag)
			delete(t.tags, tag)
		} else if value < 0 {
			log.Println("Not deleting tag that did not exist")
		} else {
			log.Println("Reducing usage count for tag", tag, "to value", value)
			t.tags[tag] = value
		}
	}

	log.Println(t.tags)
}
