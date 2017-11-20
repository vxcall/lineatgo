package lineatgo

import (
    "net/http"
    "fmt"
    "io/ioutil"
    "strings"
    "log"
    "net/url"
    "github.com/PuerkitoBio/goquery"
)

func (a *Api) getXRT(botId string)  {
    request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/userlist/",  botId), nil)
    request.Header.Set("Accept-Language", "ja")
    resp, _ := a.Client.Do(request)
    defer resp.Body.Close()
    cont, _ := ioutil.ReadAll(resp.Body)
    if strings.Index(string(cont), "XRT") == -1 {
        log.Println("Restrict Error: If you want to  keep safe to use, please use admin account!") //管理者権限でないため実行できませんでした
    }
    XRT := string(cont)[strings.Index(string(cont), "XRT") + 7:strings.Index(string(cont), "XRT") + 60]
    XRT = XRT[:strings.Index(XRT, ";") - 1]
    a.XRT = XRT
}

func (a *Api) GetAuthURL(botId string) string {
    v := url.Values{"role": {"ADMIN"}}
    request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/userlist/auth/url", botId), strings.NewReader(v.Encode()))
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
    request.Header.Set("X-CSRF-Token", a.XRT)
    resp, _ := a.Client.Do(request)
    defer resp.Body.Close()
    cont, _ := ioutil.ReadAll(resp.Body)
    return string(cont)
}

func DeleteAuthUser(XRT, botID string, LClient *http.Client) {
    request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/userlist/", botID), nil)
    request.Header.Set("Content-Type", "text/plain;charset=UTF-8")
    request.Header.Set("X-CSRF-Token", XRT)
    resp, _ := LClient.Do(request)
    defer resp.Body.Close()
    doc, _ := goquery.NewDocumentFromResponse(resp)
    doc.Find("あまね")
    /*
    "https://admin-official.line.me/%v/userlist/"に対してスクレイプ
    決裁者じゃないダミーユーザーを削除
    Content-Type text/plain;charset=UTF-8
    X-CSRF-Token
     */
}