package lineatgo

import (
    "net/http"
    "fmt"
    "strings"
    "net/url"
    "github.com/PuerkitoBio/goquery"
    "io/ioutil"
)
/*
SetName names bot
 */
func (b *Bot) SetName(newName string)  {
    v := url.Values{"role": {b.BotId}, "type": {"profile"}, "dataType": {"name"}, "name": {newName}}
    request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/account/profile/name", b.BotId), strings.NewReader(v.Encode()))
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
    request.Header.Set("X-CSRF-Token", b.api.xrt)
    response, _ := b.api.client.Do(request)
    defer response.Body.Close()
}

/*
SetStatusMessage set status message(IOW hitokoto)
 */
func (b *Bot) SetStatusMessage(newStatusMessage string)  {
    v := url.Values{"role": {b.BotId}, "type": {"profile"}, "dataType": {"hitokoto"}, "hitokoto": {newStatusMessage}}
    request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/account/profile/hitokoto", b.BotId), strings.NewReader(v.Encode()))
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
    request.Header.Set("X-CSRF-Token", b.api.xrt)
    response, _ := b.api.client.Do(request)
    defer response.Body.Close()
}

/*
GetQRCode gets qr code as byte slice
 */
func (b *Bot) GetQRCode() []byte {
    request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/account/", b.BotId), nil)
    response, _ := b.api.client.Do(request)
    defer response.Body.Close()
    doc, _ := goquery.NewDocumentFromResponse(response)
    src, _ := doc.Find("div.mdCMN08Img").Eq(0).Find("img").Attr("src")
    req, _ := http.NewRequest("GET", src, nil)
    resp, _ := b.api.client.Do(req)
    defer resp.Body.Close()
    cont, _ := ioutil.ReadAll(resp.Body)
    return cont
}

/*
GetFriendLink gets "LINE Add Link".
 */
func (b *Bot) GetFriendLink() string {
    request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/account/", b.BotId), nil)
    response, _ := b.api.client.Do(request)
    defer response.Body.Close()
    doc, _ := goquery.NewDocumentFromResponse(response)
    src, _ := doc.Find("div.mdCMN08Img").Eq(1).Find("a").Attr("href")
    return src
}

