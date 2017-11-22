package lineatgo

import (
    "net/http"
    "fmt"
    "io/ioutil"
    "strings"
    "net/url"
    "github.com/PuerkitoBio/goquery"
)

/*
GetAuthURL retrieve a url to enable access the account.
 */
func (b *Bot) GetAuthURL() string {
    v := url.Values{"role": {"ADMIN"}}
    request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/userlist/auth/url", b.BotId), strings.NewReader(v.Encode()))
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
    request.Header.Set("X-CSRF-Token", b.Api.XRT)
    resp, _ := b.Api.Client.Do(request)
    defer resp.Body.Close()
    cont, _ := ioutil.ReadAll(resp.Body)
    return string(cont)
}

type AuthUserList struct {
    Users []AuthUser
}

type AuthUser struct {
    Name string
    Delurl string
    IsPaymaster bool
    AuthorityType string
    Api *Api
}

func (b *Bot) findAuthUser() {
    request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/userlist/", b.BotId), nil)
    request.Header.Set("Content-Type", "text/plain;charset=UTF-8")
    request.Header.Set("X-CSRF-Token", b.Api.XRT)
    resp, _ := b.Api.Client.Do(request)
    defer resp.Body.Close()
    var ul []AuthUser
    doc, _ := goquery.NewDocumentFromResponse(resp)
    doc.Find("div.mdCMN08Txt").Each(func(_ int, s *goquery.Selection) {
        t := s.Find("p.mdCMN08Ttl").Text()
        u := parseAuthTxt(t)
        u.Api = b.Api
        var ok bool
        u.Delurl, ok = s.Find("div.MdM0 > input.mdBtn03Txt").Attr("data-action")
        if ok {
            ul = append(ul, u)
        }
    })
    b.AuthUserList = &AuthUserList{Users: ul}
}

func parseAuthTxt(t string) AuthUser {
    var u AuthUser
    if strings.Contains(t, "Paymaster") {
        u.IsPaymaster = true
    }
    if strings.Contains(t, "Administrator") {
        var addition int
        if u.IsPaymaster {
            addition += 13
        }
        u.Name = t[13 + addition:]
        u.AuthorityType = "Administrator"
    }
    if strings.Contains(t, "Operations personnel (no statistics view)") {
        var addition int
        if u.IsPaymaster {
            addition += 13
        }
        u.Name = t[41 + addition:]
        u.AuthorityType = "Operator(no statistics view)"
    }
    if strings.Contains(t, "Operations personnel (no authority to send)") {
        var addition int
        if u.IsPaymaster {
            addition += 13
        }
        u.Name = t[43 + addition:]
        u.AuthorityType = "Operator(no authority to send)"
    }
    return u
}

/*
DeleteAuthUser eliminate authenticated user
 */
func (u *AuthUser) Delete() {
    request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me%v", u.Delurl), nil)
    request.Header.Set("Content-Type", "text/plain;charset=UTF-8")
    request.Header.Set("X-CSRF-Token", u.Api.XRT)
    resp, _ := u.Api.Client.Do(request)
    defer resp.Body.Close()
}