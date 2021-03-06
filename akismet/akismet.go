package akismet

import (
	"net/http"
	"net/url"
	"io/ioutil"
)

func IsSpamComment(post string, address string, ip string, userAgent string, referrer string, author string, authorEmail string, apiKey string) (bool, error) {

	if len(apiKey) == 0 {
		return false, nil
	}

	values := url.Values{"blog": {address}, "user_ip": {ip}, "user_agent": {userAgent}, "referrer": {referrer}, "comment_content": {post}, "comment_author": {author}, "comment_author_email": {authorEmail}}
	resp, err := http.PostForm("http://" + apiKey + ".rest.akismet.com/1.1/comment-check", values)

	if err != nil {
		return false, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return string(body) == "true", err
}
