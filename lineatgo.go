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
    "github.com/pkg/errors"
    "math/big"
    "crypto/rsa"
    "crypto/rand"
    "github.com/PuerkitoBio/goquery"
)

type Bot struct {
    Name string
    LineId string
    BotId string
    api *Api
    AuthUserList *AuthUserList
}

type Api struct {
    mailAddress string
    password string
    client *http.Client
    xrt string
    csrfToken1 string
    csrfToken2 string
}

/*
NewApi creates a new api.
 */
func NewApi(mail, pass string) *Api {
    var api = Api{mailAddress: mail, password: pass}
    api.login()
    return &api
}

/*
NewBot creates a new bot.
 */
func (a *Api) NewBot(lineId string) (Bot, error) {
    var bot = Bot{LineId: lineId, api: a}
    err := bot.getBotInfo()
    if err != nil {
        return bot,  err
    }
    bot.getCsrfToken1()
    bot.getCsrfToken2()
    bot.findAuthUser()
    return bot, nil
}

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
    mailBox.Fill(a.mailAddress)
    passBox.Fill(a.password)
    if err := page.FindByClass("MdBtn03Login").Submit(); err != nil {
        log.Fatalf("Failed to login:%v", err)
    }

    time.Sleep(1000 * time.Millisecond)
    PINcode, err := page.FindByClass("mdLYR04PINCode").Text()
    if err != nil {
        log.Println("mailaddress ore password was wrong")
    }

    var limit bool
    ctx := context.Background()
    ctx, cancelTimer := context.WithCancel(ctx)
    fmt.Println(fmt.Sprintf("press the PINCODE below in LINE mobile: %v", PINcode))
    go timer(140000, ctx, &limit)
    for {
        title, _ := page.Title()
        if strings.Contains(title, "LINE@ MANAGER") {
            limit = false
            cancelTimer()
            break
        }

        if limit {
            log.Println("oh. timeout:(")
            limit = false
        }
    }
    c, _ := page.GetCookies()
    a.createClient(c)
    a.getXRT()
}

/*
try to login by means of http request
out of order
 */
func login2(mail, password string)  {
    jar, _ := cookiejar.New(nil)
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{ServerName: "*.line.me"},
    }
    client := &http.Client{
        Transport: tr,
        Jar: jar,
    }
    request, _ := http.NewRequest("GET", "https://admin-official.line.me/", nil)
    response, _ := client.Do(request)
    defer response.Body.Close()
    doc, _ := goquery.NewDocumentFromResponse(response)
    captchaKey, _ := doc.Find("input#captchaKey").Attr("value")
    redirectUri, _ := doc.Find("input#redirectUrl").Attr("value")
    redirectUri, _= url.QueryUnescape(redirectUri)
    state, _ := doc.Find("input#state").Attr("value")
    v, vv := getRsaKeyAndSessionKey()
    cip := rsaEncrypt(v, vv[1], mail, password)
    sendMAPW(mail, cip, vv[0], captchaKey, redirectUri, state, client)
}

/*
rsaEncrypt encrypt
 */
func rsaEncrypt(sessionKey, publicModules, mail, pass string) string {
    modInt := new(big.Int)
    modInt.SetString(publicModules, 16)
    var ex, _ = strconv.ParseInt("10001",16,0)
    var pub = rsa.PublicKey{N: modInt, E: int(ex)}

    var msg = []byte(fmt.Sprintf("%v%v %v", sessionKey, mail, pass))

    var encryption, _ = rsa.EncryptPKCS1v15(rand.Reader, &pub, msg)

    encrypted := new(big.Int)
    encrypted.SetBytes(encryption)
    return fmt.Sprintf("%x", encrypted)
}

/*
Out of order
 */
func sendMAPW(mail, cip, key, cpk, ruri, state string, client *http.Client) {
    v := url.Values{}
    v.Add("userId", mail)
    v.Add("id", key)
    v.Add("password", cip)
    v.Add("idProvider", "1")
    v.Add("response_mode", "")
    v.Add("otpId", "")
    v.Add("scope", "")
    v.Add("response_type", "code")
    v.Add("client_id", "1459630796")
    v.Add("redirect_uri", ruri)
    v.Add("displayType", "b")
    v.Add("state", state)
    v.Add("forceSecondVerification", "")
    v.Add("showPermissionApproval", "")
    v.Add("captchaKey", cpk)
    request, _ := http.NewRequest("POST", "https://access.line.me/dialog/oauth/authenticate", strings.NewReader(v.Encode()))
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse
    }
    response, _ := client.Do(request)
    defer response.Body.Close()
    fmt.Println(response.Header["Location"][0])
}

func (a *Api) createClient(c []*http.Cookie) {
    jar, _ := cookiejar.New(nil)
    u, _ := url.Parse("https://admin-official.line.me/")
    jar.SetCookies(u, c)
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{ServerName: "*.line.me"},
    }
    a.client = &http.Client{
        Transport: tr,
        Jar: jar,
    }
}

func (b *Bot) getBotInfo() error {
    request, _ := http.NewRequest("GET", "https://admin-official.line.me/api/basic/bot/list?_=1510425201579&count=10&page=1", nil)
    request.Header.Set("Content-Type", "application/json;charset=UTF-8")
    response, _ := b.api.client.Do(request)
    defer response.Body.Close()
    cont, _ := ioutil.ReadAll(response.Body)
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

/*
DeleteBot eliminates itself
 */
func (b *Bot) DeleteBot()  {
    v := url.Values{}
    v.Set("csrf_token", b.api.csrfToken2)
    v.Set("agree", "on")
    request, _ := http.NewRequest("POST", fmt.Sprintf("https://admin-official.line.me/%v/resign/", b.BotId), strings.NewReader(v.Encode()))
    request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    request.Header.Set("Upgrade-Insecure-Requests", "1")
    response, _ := b.api.client.Do(request)
    defer response.Body.Close()
}

func timer(wait int, ctx context.Context, l *bool) {
    time.Sleep(time.Duration(wait) * time.Millisecond)
    *l = true
}