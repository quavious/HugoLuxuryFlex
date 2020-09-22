package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/quavious/GoSummary/reviews"
	"github.com/quavious/GoSummary/translate"
)

func main() {
	_csv, err := os.Open("crafts.csv")
	if err != nil {
		panic(err)
	}
	_reader := csv.NewReader(_csv)
	//flag = true

	oneRecord, err := _reader.ReadAll()
	if err == io.EOF {
		log.Fatalln(err)
		panic(err)
	}
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	for _, id := range oneRecord[:15] {
		fmt.Println(id[0])
		productURL := "https://www.amazon.com/dp/" + id[0] + "?dchild=1&qid=1599296775&s=hi"
		amazonURL := "https://www.amazon.com/product-reviews/" + id[0] + "/ref=cm_cr_unknown?filterByStar=five_star&reviewerType=all_reviews&pageNumber=1"
		_amazonURL := "https://www.amazon.com/product-reviews/" + id[0] + "/ref=cm_cr_unknown?filterByStar=one_star&reviewerType=all_reviews&pageNumber=1"
		options := []chromedp.ExecAllocatorOption{
			chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.102 Safari/537.36 Edg/85.0.564.51"),
			chromedp.Flag("Headless", false),
		}
		ctx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
		defer cancel()

		ctx, cancel = chromedp.NewContext(ctx)
		defer cancel()

		_ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
		defer cancel()

		var _html string
		if err := chromedp.Run(_ctx,
			chromedp.Navigate(productURL),
			chromedp.Sleep(time.Second*3),
			chromedp.OuterHTML("html", &_html),
		); err != nil {
			panic(err)
		}
		_doc, _ := goquery.NewDocumentFromReader(strings.NewReader(_html))

		title := strings.TrimSpace(_doc.Find("#productTitle").Text())
		price := strings.TrimSpace(_doc.Find("#priceblock_ourprice").Text())
		category := ""

		_doc.Find(".a-link-normal.a-color-tertiary").Each(func(i int, s *goquery.Selection) {
			if i == 0 {
				category = strings.TrimSpace(s.Text())
			}
		})

		poster, _ := _doc.Find("#landingImage").Attr("src")
		description := ""
		index := 0
		_doc.Find(".a-unordered-list.a-vertical.a-spacing-mini").Find("span").Each(func(i int, s *goquery.Selection) {
			scraped := strings.TrimSpace(s.Text())
			index++
			description += scraped + ". "
		})
		_doc.Find("#productDescription").Find("p").Each(func(i int, s *goquery.Selection) {
			scraped := strings.TrimSpace(s.Text())
			if len(scraped) != 0 {
				index++
				description += scraped + ".\n"
			}
		})
		var goodReview string
		var goodContent string
		if err := chromedp.Run(_ctx,
			chromedp.Navigate(amazonURL),
			chromedp.Sleep(time.Second*3),
			chromedp.OuterHTML("html", &goodReview),
		); err != nil {
			panic(err)
		}
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(goodReview))
		doc.Find(".a-size-base.review-text.review-text-content").Each(func(i int, sel *goquery.Selection) {
			temp := strings.TrimSpace(sel.Text())
			if len(temp) < 800 {
				temp, _ = translate.Convert("ko", temp)
				time.Sleep(5 * time.Second)
				temp, _ = translate.Convert("en", temp)
				time.Sleep(5 * time.Second)
				goodContent += temp + " "
			}
		})

		var badReview string
		var badContent string
		if err := chromedp.Run(ctx,
			chromedp.Navigate(_amazonURL),
			chromedp.Sleep(time.Second*3),
			chromedp.OuterHTML("html", &badReview),
		); err != nil {
			panic(err)
		}
		doc, _ = goquery.NewDocumentFromReader(strings.NewReader(badReview))
		doc.Find(".a-size-base.review-text.review-text-content").Each(func(i int, sel *goquery.Selection) {
			temp := strings.TrimSpace(sel.Text())
			if len(temp) < 800 {
				temp, _ = translate.Convert("ko", temp)
				time.Sleep(5 * time.Second)
				temp, _ = translate.Convert("en", temp)
				time.Sleep(5 * time.Second)
				badContent += temp + " "
			}
		})

		spec := reviews.Summarize(description, (index)/2)
		var s string
		fmt.Println(spec)
		for _, item := range spec {
			s += item + " "
		}

		time.Sleep(time.Second * 10)
		good := reviews.Summarize(goodContent, 5)

		bad := reviews.Summarize(badContent, 5)

		if len(strings.TrimSpace(price)) == 0 {
			price = "Need To Check"
		}

		err = os.Chdir("C:/Hugo/amazon")
		if err != nil {
			log.Fatalln(err)
			chromedp.Cancel(_ctx)
			return
		}
		fmt.Println(os.Getwd())
		path := time.Now().Format("2006\\01\\02\\15")
		filename := time.Now().Format("20060102150405") + ".md"
		cmd := exec.Command("hugo", "new", "posts\\"+path+"\\"+filename)
		cmd.Dir = "C:\\Hugo\\amazon"

		output, err := cmd.Output()

		if err != nil {
			log.Fatalln(err)
			chromedp.Cancel(_ctx)
			return
		}
		fmt.Println(string(output))

		var postContent string

		postContent += "---\n"
		postContent += "title : " + title + "\n"
		postContent += "date : " + time.Now().Format("2006-01-02T15:04:05") + "+0900" + "\n"
		postContent += "draft : false\n"
		postContent += "image : " + poster + "\n"
		postContent += "price : " + price + "\n"
		postContent += "text : " + s + "\n"
		postContent += "categories : " + category + "\n"
		postContent += "good : \n"
		if len(good) != 0 {
			for _, g := range good {
				postContent += " - " + g + "\n"
			}
		} else {
			postContent += " - " + "Nothing Pros... or Try checking this item directly" + "\n"
		}
		postContent += "bad : \n"

		if len(bad) != 0 {
			for _, b := range bad {
				postContent += " - " + b + "\n"
			}
		} else {
			postContent += " - " + "Nothing Cons... or Try checking this item directly" + "\n"
		}
		postContent += "productID : " + id[0] + "\n"
		postContent += "---\n\n"

		err = ioutil.WriteFile("C:\\Hugo\\amazon\\content\\posts\\"+path+"\\"+filename, []byte(postContent), 0644)

		if err != nil {
			log.Fatalln(err)
			chromedp.Cancel(_ctx)
			return
		}
		interval := rand.Intn(60) + 60

		fmt.Println("Make File Successful")
		time.Sleep(time.Second * time.Duration(interval))
	}
}
