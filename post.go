package lineatgo

import (
    "github.com/PuerkitoBio/goquery"
    "fmt"
    "log"
    "net/url"
    "net/http"
    "strings"
    "time"
)

/*
DeletePostAll delete all of post the account has.
 */
func (a *Api) DeletePostAll(botId string) {
    request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/home/", botId), nil)
    request.Header.Set("Accept-Language", "ja")
    resp, _ := a.Client.Do(request)
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromResponse(resp)
    if err != nil {
        log.Fatalf("create document error: %v", err)
    }
    a.getCsrfToken(doc)
    go a.retrievePost(doc, botId)
    time.Sleep(10*time.Second)
}

func (a *Api) getCsrfToken(doc *goquery.Document) {
    s := doc.Find("script#postEditForm\\.html").Text()

    doc2, err := goquery.NewDocumentFromReader(strings.NewReader(s))
    if err != nil {
        log.Fatalf("create document error: %v", err)
    }
    a.CsrfToken,  _ = doc2.Find("#postForm > input").First().Attr("value")
}

func (a *Api) retrievePost(doc *goquery.Document, botId string) {
    doc.Find("div.mdCMN13Foot > a").Each(func(_ int, s *goquery.Selection) {
        url, _ := s.Attr("href")
        deluri := fmt.Sprintf("https://admin-official.line.me/%v/home/%v/delete", botId,  url[2:len(url) - 9])
        go a.postDel(deluri)
    })
    l, ok := doc.Find("a.nextLink").Attr("href")
    if ok {
        go a.fetchNextPage(fmt.Sprintf("https://admin-official.line.me/%v/home/%v", botId, l), botId)
    } else {
        fmt.Println("deleted all posts")
    }
}

func (a *Api) fetchNextPage(uri string, botId string) {
    request, _ := http.NewRequest("GET", uri, nil)
    request.Header.Set("Accept-Language", "ja")
    resp, _ := a.Client.Do(request)
    defer resp.Body.Close()
    doc, err := goquery.NewDocumentFromResponse(resp)
    if err != nil {
        log.Fatalf("create document error: %v", err)
    }
    go a.retrievePost(doc, botId)
}

func (a *Api) postDel(uri string)  {
    v := url.Values{"csrf_token": {a.CsrfToken}}
    request, _ := http.NewRequest("POST", uri, strings.NewReader(v.Encode()))
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
    resp, err := a.Client.Do(request)
    if err != nil {
        log.Fatalf("DELETE ERROR: %v", err)
    }
    defer resp.Body.Close()
}
