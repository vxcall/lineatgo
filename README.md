[![GoDoc](https://godoc.org/github.com/s3pt3mb3r/lineatgo?status.svg)](https://godoc.org/github.com/s3pt3mb3r/lineatgo)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)
# lineatgo
This is a unofficial LINE@ API that was implemented in pure go

## installation
```
go get github.com/s3pt3mb3r/lineatgo
```

## warning
This library doesn't handle errors yet.
I'll do it in a few days.

## Usage
```go
package main

import (
    "github.com/s3pt3mb3r/lineatgo"
    "log"
    "fmt"
)

func main()  {
    api := lineatgo.NewApi("MAIL_ADDRESS", "PASSWORD")
    bot, err := api.NewBot("@LINEID")
    if err != nil {
        log.Println(err)
    }
    url := bot.GetAuthURL(lineatgo.Administrator)
    //what else: lineatgo.Operator, lineatgo.LimitedOperator, lineatgo.Messenger
    fmt.Println(url)

    qr := bot.GetQRCode()
    file, err := os.OpenFile("test.png", os.O_RDWR|os.O_CREATE, 0666)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    file.Write(qr)
}
```

## Todo
- [x] Enable to select authority type in getAuth function
- [x] Enable to Delete paymaster user's clearance
- [x] Enable to Post some text on time line
- [ ] Enable to Post image or video on time line
- [ ] Fix DeletePostAll function

### At last
Probably, being overlook some factors, I can't code Login() function without web driver

So If it's possible, please make Login() function more better and send pull request:)