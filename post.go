package lineatgo

import (
    "github.com/PuerkitoBio/goquery"
    "fmt"
    "log"
    "net/url"
    "net/http"
    "strings"
)

/*
DeletePostAll delete all of post the account has.
 */
func (b *Bot) DeletePostAll() {
    request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/home/", b.BotId), nil)
    response, _ := b.Api.Client.Do(request)
    defer response.Body.Close()

    doc, err := goquery.NewDocumentFromResponse(response)
    if err != nil {
        log.Fatalf("create document error: %v", err)
    }

    var endChan = make(chan bool)
    go b.retrievePost(doc, endChan)
    notice := <- endChan
    fmt.Println(notice)
}

func (b *Bot) retrievePost(doc *goquery.Document, endChan chan bool) {
    doc.Find("div.mdCMN13Foot > a").Each(func(_ int, s *goquery.Selection) {
        url, _ := s.Attr("href")
        deluri := fmt.Sprintf("https://admin-official.line.me/%v/home/%v/delete", b.BotId,  url[2:len(url) - 9])
        go b.postDel(deluri, endChan)
    })
    l, ok := doc.Find("a.nextLink").Attr("href")
    if ok {
        go func() {
            request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/home/%v", b.BotId, l), nil)
            response, _ := b.Api.Client.Do(request)
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
    v := url.Values{"csrf_token": {b.Api.CsrfToken}}
    request, _ := http.NewRequest("POST", uri, strings.NewReader(v.Encode()))
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
    response, _ := b.Api.Client.Do(request)
    defer response.Body.Close()
    endChan <- true
}

func (b *Bot) PostText(text string) {
    v := url.Values{}
    v.Set("csrf_token", b.Api.CsrfToken)
    v.Set("scheduled", "")
    v.Set("sendDate", "")
    v.Set("sendHour", "0")
    v.Set("minutes1", "0")
    v.Set("minutes2", "0")
    v.Set("sendTimeType", "NOW")
    v.Set("contentType1", "TEXT")
    v.Set("body", text)
    v.Set("draftId", "")
    request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/home/api/posts", b.BotId), strings.NewReader(v.Encode()))
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
    response, _ := b.Api.Client.Do(request)
    defer response.Body.Close()
}