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

package main

import (
    "fmt"
    "flag"
    "log"
    "os"

    "github.com/mattn/go-sqlite3"
    "github.com/bwmarrin/discordgo"
    "github.com/ttamre/go.watchlist/bot"
)

const DEFAULT_DB_PATH = "data/database.db"


func main() {
    // Process command line flags
    db_file := flag.String("database", DEFAULT_DB_PATH, "database file path")
    flag.Parse()

    // Creating a session to connect to discord server
    session, err := discordgo.New("Bot " + os.Getenv("DISCORD_WATCHLIST_BOT_TOKEN"))
    if err != nil {
        log.Fatal(err)
    }

    // Creating a database connection
    db, err := sql.Open("sqlite3", db_file)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Registering handlers
    session.AddHandler(bot.MasterHandler)

    // Open a websocket connection to Discord and begin listening.
    err = session.Open()
    if err != nil {
        fmt.Println("Error opening connection: ", err)
        return
    }
    defer session.Close()

    // Simple way to keep program running until CTRL-C is pressed
    fmt.Println("bot is now running, press ctrl-c to exit...")
    <-make(chan struct{})
}
