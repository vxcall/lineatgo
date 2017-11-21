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
    "fmt"
)

func main() {
    api := lineatgo.NewApi("example@mail.com", "passw0rd")

    api.Login()
    botId := api.GetBotIdByName("BOT_NAME")

    api.DeletePostAll(botId) //delete all posts
    api.GetAuthURL(botId) //get authority URL
}
```

### At last
Probably, being overlook some factors, I can't code Login() function without web driver

So If it's possible, please make Login() function more better and send pull request:)