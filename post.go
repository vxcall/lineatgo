package lineatgo

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

/*
DeletePostAll deletes all of post the account has.
*/
func (b *Bot) DeletePostAll() {
	request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/home/", b.BotId), nil)
	response, _ := b.api.client.Do(request)
	defer response.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		log.Fatalf("create document error: %v", err)
	}

	var endChan = make(chan bool)
	go b.retrievePost(doc, endChan)
	notice := <-endChan
	fmt.Println(notice)
}

func (b *Bot) retrievePost(doc *goquery.Document, endChan chan bool) {
	doc.Find("div.mdCMN13Foot > a").Each(func(_ int, s *goquery.Selection) {
		url, _ := s.Attr("href")
		deluri := fmt.Sprintf("https://admin-official.line.me/%v/home/%v/delete", b.BotId, url[2:len(url)-9])
		go b.postDel(deluri, endChan)
	})
	l, ok := doc.Find("a.nextLink").Attr("href")
	if ok {
		go func() {
			request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/home/%v", b.BotId, l), nil)
			response, _ := b.api.client.Do(request)
			defer response.Body.Close()
			doc, err := goquery.NewDocumentFromResponse(response)
			if err != nil {
				log.Fatalf("create document error: %v", err)
			}
			go b.retrievePost(doc, endChan)
		}()
	}
}

func (b *Bot) postDel(uri string, endChan chan bool) {
	v := url.Values{"csrf_token": {b.api.csrfToken1}}
	request, _ := http.NewRequest("POST", uri, strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	response, _ := b.api.client.Do(request)
	defer response.Body.Close()
	endChan <- true
}

/*
PostText makes it possible to post composed of text only
*/
func (b *Bot) PostText(text string) {
	v := url.Values{}
	v.Set("csrf_token", b.api.csrfToken1)
	v.Set("body", text)

	request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/home/api/posts", b.BotId), strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	response, _ := b.api.client.Do(request)
	defer response.Body.Close()
}

type Post struct {
	imagePath1 string
	imagePath2 string
	imagePath3 string
	imagePath4 string
}

func (b *Bot) PostWithImages(text string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	f, err := os.Open("/Users/am4ne/Pictures/white_kabekin.jpg")
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	fw, err := w.CreateFormFile("file", "/Users/am4ne/Pictures/white_kabekin.jpg")
	if err != nil {
		log.Println(err)
	}
	if _, err = io.Copy(fw, f); err != nil {
		log.Println(err)
	}
	w.Close()

	v := url.Values{"file": {buf.String()}, "csrf_token": {b.api.csrfToken1}}

	request, err := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/home/api/objects", b.BotId), strings.NewReader(v.Encode()))
	if err != nil {
		log.Println(err)
	}
	request.Header.Set("Content-Type", w.FormDataContentType())
	request.Header.Set("Referer", "https://admin-official.line.me/9869462/home/send/")
	response, err := b.api.client.Do(request)
	if err != nil {
		log.Println(err)
	}
	defer response.Body.Close()
	fmt.Println(response.StatusCode)
	cont, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(cont))
}

func (b *Bot) SendFirstStage() {
	file, err := os.Open("/Users/am4ne/Pictures/3398280846-white_kabekin-mw4R-1440x900-MM-100.jpg")
	if err != nil {
		// Openエラー処理
	}
	defer file.Close()
	v := url.Values{}
	var output string
	file.Write([]byte(output))
	v.Set("file", output)
	v.Set("csrf_token", b.api.csrfToken1)
	request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/home/api/posts", b.BotId), strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundaryuPiicl3hB2rPuzwJ")
	response, _ := b.api.client.Do(request)
	defer response.Body.Close()
	cont, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(cont))
}
