package lineatgo

import (
    "net/http"
    "time"
    "fmt"
    "log"
    "strings"
    "context"
    "github.com/sclevine/agouti"
    "net/http/cookiejar"
    "net/url"
    "crypto/tls"
    "io/ioutil"
    "encoding/json"
    "github.com/mattn/go-scan"
    "strconv"
    "github.com/PuerkitoBio/goquery"
    "github.com/pkg/errors"
)

type Bot struct {
    Name string
    LineId string
    BotId string
    Api *Api
    AuthUserList *AuthUserList
}

type Api struct {
    MailAddress string
    Password string
    Client *http.Client
    XRT string
    CsrfToken string
}

/*
NewApi create a new api.
 */
func NewApi(mail, pass string) *Api {
    var api = Api{MailAddress: mail, Password: pass}
    api.login()
    return &api
}

/*
NewBot create a new bot.
 */
func (a *Api) NewBot(lineId string) (Bot, error) {
    var bot = Bot{LineId: lineId, Api: a}
    err := bot.getBotInfo()
    if err != nil {
        return bot,  err
    }
    bot.getCsrfToken()
    bot.findAuthUser()
    return bot, nil
}

/*
Login log in account using mail address and password
 */
func (a *Api) login() {
    driver := agouti.ChromeDriver(agouti.ChromeOptions("args", []string{"--headless", "--disable-gpu"}), )
    if err := driver.Start(); err != nil {
        log.Fatalf("Failed to start driver:%v", err)
    }
    defer driver.Stop()

    page, err := driver.NewPage()
    if err != nil {
        log.Fatalf("Failed to open page:%v", err)
    }

    if err := page.Navigate("https://admin-official.line.me/"); err != nil {
        log.Fatalf("Failed to navigate:%v", err)
    }

    mailBox := page.FindByID("id")
    passBox := page.FindByID("passwd")
    mailBox.Fill(a.MailAddress)
    passBox.Fill(a.Password)
    if err := page.FindByClass("MdBtn03Login").Submit(); err != nil {
        log.Fatalf("Failed to login:%v", err)
    }

    time.Sleep(1000 * time.Millisecond)
    PINcode, err := page.FindByClass("mdLYR04PINCode").Text()
    if err != nil {
        log.Println("メールアドレスまたはパスワードが間違っています。")
    }

    var limit bool
    ctx := context.Background()
    ctx, cancelTimer := context.WithCancel(ctx)
    fmt.Println(fmt.Sprintf("携帯のLINEで以下のPINコードを入力してください: %v", PINcode))
    go timer(140000, ctx, &limit)
    for {
        title, _ := page.Title()
        if strings.Contains(title, "LINE@ MANAGER") {
            limit = false
            cancelTimer()
            break
        }

        if limit {
            log.Println("時間切れです。")
            limit = false
        }
    }
    c, _ := page.GetCookies()
    a.createClient(c)
    a.getXRT()
}

func (a *Api) createClient(c []*http.Cookie) {
    jar, _ := cookiejar.New(nil)
    u, _ := url.Parse("https://admin-official.line.me/")
    jar.SetCookies(u, c)
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{ServerName: "*.line.me"},
    }
    a.Client = &http.Client{
        Transport: tr,
        Jar: jar,
    }
}

func (b *Bot) getBotInfo() error {
    request, _ := http.NewRequest("GET", "https://admin-official.line.me/api/basic/bot/list?_=1510425201579&count=10&page=1", nil)
    request.Header.Set("Content-Type", "application/json;charset=UTF-8")
    resp, _ := b.Api.Client.Do(request)
    defer resp.Body.Close()
    cont, _ := ioutil.ReadAll(resp.Body)
    var ij interface{}
    json.Unmarshal([]byte(string(cont)), &ij)
    var (
        displayName string
        lineId string
        botId int
    )
    for i:=0; i<strings.Count(string(cont), "botId"); i++ {
        scan.ScanTree(ij, fmt.Sprintf("/list[%v]/lineId", i), &lineId)
        if lineId != b.LineId {
            continue
        }
        scan.ScanTree(ij, fmt.Sprintf("/list[%v]/displayName", i), &displayName)
        scan.ScanTree(ij, fmt.Sprintf("/list[%v]/botId", i), &botId)
        b.Name = displayName
        b.BotId = strconv.Itoa(botId)
        break
    }
    if b.Name == "" {
        return errors.New(fmt.Sprintf(`ERROR: Specified bot "%v" was not found.\nUse other LINE account and try again:)`, b.LineId))
    }
    return nil
}

func (a *Api) getXRT()  {
    request, _ := http.NewRequest("GET", "https://admin-official.line.me/", nil)
    request.Header.Set("Accept-Language", "ja")
    resp, _ := a.Client.Do(request)
    defer resp.Body.Close()
    cont, _ := ioutil.ReadAll(resp.Body)
    XRT := string(cont)[strings.Index(string(cont), "XRT") + 7:strings.Index(string(cont), "XRT") + 60]
    XRT = XRT[:strings.Index(XRT, ";") - 1]
    a.XRT = XRT
}

func (b *Bot) getCsrfToken() {
    request, _ := http.NewRequest("GET", fmt.Sprintf("https://admin-official.line.me/%v/home/", b.BotId), nil)
    request.Header.Set("Accept-Language", "ja")
    resp, _ := b.Api.Client.Do(request)
    defer resp.Body.Close()

    doc, err := goquery.NewDocumentFromResponse(resp)
    if err != nil {
        log.Fatalf("create document error: %v", err)
    }
    s := doc.Find("script#postEditForm\\.html").Text()

    doc2, err := goquery.NewDocumentFromReader(strings.NewReader(s))
    if err != nil {
        log.Fatalf("create document error: %v", err)
    }
    b.Api.CsrfToken,  _ = doc2.Find("#postForm > input").First().Attr("value")
}

func timer(wait int, ctx context.Context, l *bool) {
    time.Sleep(time.Duration(wait) * time.Millisecond)
    *l = true
}