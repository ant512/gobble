package main

type BlogPost struct {
	title string
	publishDate string
	tags []string
	body string
}

func (b *BlogPost) SetTitle(s string) {
	b.title = s
}

func (b *BlogPost) Title() string {
	return b.title
}

func (b *BlogPost) SetPublishDate(s string) {
	b.publishDate = s
}

func (b *BlogPost) PublishDate() string {
	return b.publishDate
}

func (b *BlogPost) SetTags(s []string) {
	b.tags = s
}

func (b *BlogPost) Tags() []string {
	return b.tags
}

func (b *BlogPost) SetBody(s string) {
	b.body = s
}

func (b *BlogPost) Body() string {
	return b.body
}
