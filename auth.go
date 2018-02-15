package lineatgo

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const (
	Administrator   = "ADMIN"
	Operator        = "OPERATOR"
	LimitedOperator = "OPERATOR_LIMITED"
	Messenger       = "MESSENGER"
)

/*
GetAuthURL retrieves a url to enable access the account.
*/
func (b *bot) GetAuthURL(role string) string {
	v := url.Values{"role": {role}}
	request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/userlist/auth/url", b.BotId), strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	request.Header.Set("X-CSRF-Token", b.xrt)
	response, _ := b.client.Do(request)
	defer response.Body.Close()
	cont, _ := ioutil.ReadAll(response.Body)
	return string(cont)
}

type AuthUserList struct {
	Users []AuthUser
}

type AuthUser struct {
	Name          string
	id            string
	botId         string
	IsPaymaster   bool
	AuthorityType string
	*api
}

func (b *bot) findAuthUser() {
	request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/userlist/", b.BotId), nil)
	request.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	request.Header.Set("X-CSRF-Token", b.xrt)
	response, _ := b.client.Do(request)
	defer response.Body.Close()
	var ul []AuthUser
	doc, _ := goquery.NewDocumentFromResponse(response)
	doc.Find("div.MdCMN08ImgSet").Each(func(_ int, s *goquery.Selection) {
		t := s.Find("p.mdCMN08Ttl").Text()
		u := parseAuthTxt(t, AuthUser{api: b.api})
		u.botId = b.BotId
		imgurl, _ := s.Find("div.mdCMN08Img > img").Attr("src")
		u.id = imgurl[len(fmt.Sprintf("/%v/userlist/profile/", b.BotId)):]
		ul = append(ul, u)
	})
	b.AuthUserList = &AuthUserList{Users: ul}
}

func parseAuthTxt(t string, u AuthUser) AuthUser {
	if strings.Contains(t, "Paymaster") {
		u.IsPaymaster = true
	}
	if strings.Contains(t, "Administrator") {
		var addition int
		if u.IsPaymaster {
			addition += 13
		}
		u.Name = t[13+addition:]
		u.AuthorityType = "Administrator"
	}
	if strings.Contains(t, "Operations personnel (no statistics view)") {
		var addition int
		if u.IsPaymaster {
			addition += 13
		}
		u.Name = t[41+addition:]
		u.AuthorityType = "Operator(no statistics view)"
	}
	if strings.Contains(t, "Operations personnel (no authority to send)") {
		var addition int
		if u.IsPaymaster {
			addition += 13
		}
		u.Name = t[43+addition:]
		u.AuthorityType = "Operator(no authority to send)"
	}
	return u
}

/*
DeleteAuthUser purges user from bot
*/
func (u *AuthUser) Delete() error {
	if u.IsPaymaster {
		return errors.New("ERROR: This user is a paymaster. Please execute SetPaymaster to other user.")
	}
	delurl := fmt.Sprintf("/%v/userlist/del/%v", u.botId, u.id)
	request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me%v", delurl), nil)
	request.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	request.Header.Set("X-CSRF-Token", u.xrt)
	response, _ := u.client.Do(request)
	defer response.Body.Close()
	return nil
}

/*
SetPaymaster changes payer for this bot
*/
func (u AuthUser) SetPaymaster() {
	request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/userlist/api/users/payperson/%v", u.botId, u.id), nil)
	request.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	request.Header.Set("X-CSRF-Token", u.xrt)
	response, _ := u.client.Do(request)
	defer response.Body.Close()
}
