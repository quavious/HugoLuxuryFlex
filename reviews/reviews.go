package reviews

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fedesog/webdriver"
	"github.com/quavious/GoSummary/credential"
)

type container struct {
	Result []string `json:"sentences"`
}

func (c *container) filter() []string {
	temp := []string{}
	dic := map[string]int{}
	for _, t := range c.Result {
		if dic[t] != 1 {
			temp = append(temp, t)
			dic[t] = 1
		}
	}
	c = &container{}
	return temp
}

//Summarize summaries very large string.
func Summarize(content string, items int) []string {
	container := container{}
	url := "https://aylien-text.p.rapidapi.com/summarize"
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("x-rapidapi-host", "aylien-text.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", credential.SummarizeKEY)

	q := req.URL.Query()
	q.Add("title", "ProductReview")
	q.Add("text", content)
	//q.Add("mode", "short")
	q.Add("sentences_number", strconv.Itoa(items))
	req.URL.RawQuery = q.Encode()
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &container)

	elements := container.filter()
	return elements
}

//ReturnReviews gives you good and bad review.
func ReturnReviews(url string, sess *webdriver.Session) (string, error) {
	err := sess.Url(url)
	if err != nil {
		return "", errors.New("scraping failed")
	}
	source, _ := sess.Source()
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(source))
	var content string

	doc.Find(".a-size-base.review-text.review-text-content").Each(func(i int, sel *goquery.Selection) {
		temp := strings.TrimSpace(sel.Text())
		content += temp + " \n"
	})
	return content, nil
}
