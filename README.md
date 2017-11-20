# lineatgo
LINE@ API implementation in pure Go

# Usage
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