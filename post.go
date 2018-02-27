package lineatgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

/*
DeletePostAll deletes all of post the account has.
*/
func (b *Bot) DeletePostAll() {
	request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/home/", b.BotId), nil)
	response, _ := b.client.Do(request)
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
			response, _ := b.client.Do(request)
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
	v := url.Values{"csrf_token": {b.csrfToken1}}
	request, _ := http.NewRequest("POST", uri, strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	response, _ := b.client.Do(request)
	defer response.Body.Close()
	endChan <- true
}

type Post struct {
	Images []string
	Text   string
	*Api
	*Bot
}

/*
NewPost initialize post struct which can be added component
*/
func (b *Bot) NewPost() *Post {
	return &Post{Api: b.Api, Bot: b}
}

/*
Add add post component, like text or images
*/
func (p *Post) Add(category string, content ...string) {
	switch category {
	case "image":
		p.Images = append(p.Images, content...)
		if len(p.Images) > 9 {
			p.Images = p.Images[:9]
		}
	case "text":
		if p.Text != "" {
			p.Text = p.Text + ("\n" + strings.Join(content, ""))
		} else {
			p.Text = strings.Join(content, "")
		}
	}
}

/*
Post makes it possible to post composed of text and images(photos videos)
*/
func (p *Post) Post() {
	var comp []imageData
	count := len(p.Images)
	for _, i := range p.Images {
		imd := p.getObjectData(i)
		comp = append(comp, imd)
	}
	for i := 0; i <= 8-count; i++ {
		var u imageData
		u.Media.Type = "PHOTO"
		u.Media.Width = 0
		u.Media.Height = 0
		u.Media.ObjectId = ""
		comp = append(comp, u)
	}
	request := p.customizeReq(comp)
	response, err := p.client.Do(request)
	if err != nil {
		log.Println(err)
	}
	defer response.Body.Close()
}

func (p *Post) customizeReq(comp []imageData) *http.Request {
	v := url.Values{"csrf_token": {p.csrfToken1}, "scheduled": {""}, "tzOffset": {"-540"}, "sendDate": {""}, "sendHour": {"0"}, "minutes1": {"0"}, "minutes2": {"0"}, "sendTimeType": {"NOW"}, "contentType1": {"MULTI_IMAGE"}, "draftId": {""}}
	v.Set("body", p.Text)

	for i := 0; i <= 8; i++ {
		if comp[i].Media.ObjectId == "" {
			v.Set(fmt.Sprintf("media[%v].objectId", strconv.Itoa(i)), "")
			v.Set(fmt.Sprintf("media[%v].type", strconv.Itoa(i)), "PHOTO")
			v.Set(fmt.Sprintf("media[%v].width", strconv.Itoa(i)), "")
			v.Set(fmt.Sprintf("media[%v].height", strconv.Itoa(i)), "")
		} else {
			v.Set(fmt.Sprintf("media[%v].objectId", strconv.Itoa(i)), comp[i].Media.ObjectId)
			v.Set(fmt.Sprintf("media[%v].type", strconv.Itoa(i)), comp[i].Media.Type)
			v.Set(fmt.Sprintf("media[%v].width", strconv.Itoa(i)), strconv.Itoa(comp[i].Media.Width))
			v.Set(fmt.Sprintf("media[%v].height", strconv.Itoa(i)), strconv.Itoa(comp[i].Media.Height))
		}
	}
	request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/home/api/posts", p.BotId), strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	return request
}

type imageData struct {
	Media struct {
		Type     string `json:"type"`
		Height   int    `json:"height"`
		Width    int    `json:"width"`
		ObjectId string `json:"objectId"`
	} `json:"media"`
}

func (b *Bot) getObjectData(path string) imageData {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	f, err := os.Open(path)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	fw, err := w.CreateFormFile("file", path)
	if err != nil {
		log.Println(err)
	}
	if _, err = io.Copy(fw, f); err != nil {
		log.Println(err)
	}
	w.WriteField("csrf_token", b.csrfToken1)
	w.Close()

	request, err := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/home/api/objects", b.BotId), &buf)
	if err != nil {
		log.Println(err)
	}
	request.Header.Set("Content-Type", w.FormDataContentType())
	response, err := b.client.Do(request)
	if err != nil {
		log.Println(err)
	}
	defer response.Body.Close()
	cont, _ := ioutil.ReadAll(response.Body)
	var d imageData
	if err := json.Unmarshal(cont, &d); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return d
	}
	return d
}
