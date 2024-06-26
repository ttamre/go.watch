/*
go.watchlist - a watchlist manager discord bot
Copyright (C) 2024 Tem Tamre

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
    "fmt"
    "flag"
    "log"
    "os"

    "github.com/go-redis/redis/v8"
    "github.com/bwmarrin/discordgo"
    "github.com/ttamre/go.watchlist/bot"
)


func main() {
    // Process command line flags
    var (
        redis_host = flag.String("host", "localhost", "Redis host")
        redis_port = flag.String("port", "3001", "Redis port")
    )
    flag.Parse()

    // Creating a session to connect to discord server
    session, err := discordgo.New("Bot " + os.Getenv("DISCORD_WATCHLIST_BOT_TOKEN"))
    if err != nil {
        log.Fatal(err)
    }

    // Creating a database connection
    rdb := redis.NewClient(&redis.Options{
        Addr:       *redis_host + ":" + *redis_port,
        Password:   "",
        DB:         0,
    }

    /*
    session.AddHandler(bot.AddHandler(&rdb))
    session.AddHandler(bot.DeleteHandler(&rdb))
    session.AddHandler(bot.ViewHandler(&rdb))
    session.AddHandler(bot.UpdateHandler(&rdb))
    session.AddHandler(bot.HelpHandler(&rdb))
    */

    // Open a websocket connection to Discord and begin listening.
    err = session.Open()
    if err != nil {
        fmt.Println("Error opening connection: ", err)
        return
    }

    // Simple way to keep program running until CTRL-C is pressed
    fmt.Println("Bot is now running, press CTRL-C to exit...")
    <-make(chan struct{})
}
