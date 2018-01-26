package lineatgo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func (a *Api) getXRT() {
	request, _ := http.NewRequest("GET", "https://admin-official.line.me/", nil)
	response, _ := a.client.Do(request)
	defer response.Body.Close()
	cont, _ := ioutil.ReadAll(response.Body)
	xrt := regexp.MustCompile(`var XRT = '[a-zA-Z0-9]+'`).FindString(string(cont))
	a.xrt = xrt[11 : len(xrt)-1]
}

func (b *Bot) getCsrfToken1() {
	request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/home/", b.BotId), nil)
	response, _ := b.api.client.Do(request)
	defer response.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		log.Fatalf("create document error: %v", err)
	}
	s := doc.Find("script#postEditForm\\.html").Text()

	doc2, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		log.Fatalf("create document error: %v", err)
	}
	b.api.csrfToken1, _ = doc2.Find("#postForm > input").First().Attr("value")
}

func (b *Bot) getCsrfToken2() {
	request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/resign/", b.BotId), nil)
	response, _ := b.api.client.Do(request)
	defer response.Body.Close()
	doc, _ := goquery.NewDocumentFromResponse(response)
	b.api.csrfToken2, _ = doc.Find("form > input").Attr("value")
}

type LKey struct {
	RsaKey     string `json: "rsa_key"`
	SessionKey string `json: "session_key"`
}

func getRsaKeyAndSessionKey() (string, []string) {
	client := &http.Client{}
	unixTime := time.Now().Local().UnixNano()
	us := strconv.FormatInt(unixTime, 10)
	v := url.Values{"_": {us[:len(us)-6]}}
	req, _ := http.NewRequest("GET", "https://access.line.me/authct/v1/keys/line", nil)
	req.Header.Set("Referer", "https://access.line.me/")
	req.URL.RawQuery = v.Encode()
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	cont, _ := ioutil.ReadAll(resp.Body)
	var lkey struct {
		SessionKey string `json:"session_key"`
		RsaKey     string `json:"rsa_key"`
	}
	if err := json.Unmarshal(cont, &lkey); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
	}
	return lkey.SessionKey, strings.Split(lkey.RsaKey, ",")
}
