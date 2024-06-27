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

package bot

import (
    "fmt"
    "math/rand"

    "github.com/mattn/go-sqlite3"
)

const TIME_FORMAT = "2024-01-01 00:00:00"

// Entry represents a single entry in the watchlist
type Entry struct {
    UserID      string      `json:"user_id"`
    Title       string      `json:"title"`
    Category    Category    `json:"category"`
    Date        time.Time   `json:"date"`
    Link        string      `json:"link"`
}

// Category represents the type of item in the watchlist
type Category string
const (
    Movie   Category = "movie"
    Show    Category = "show"
    Anime   Category = "anime"
)

// Load the entry from the database
func GetEntryFromDB(db *sql.DB, userID string, title string) *Entry {
    // Load the entry from the database
    // Return the entry
    //
    var e Entry
    err := db.QueryRow("SELECT userID, title, category, date, link" +
                "FROM entries WHERE userID = ? and title = ?", userID, title)
    err = err.Scan(&e.UserID, &e.Title, &e.Category, &e.Date, &e.Link)
    if err != nil {
        log.Fatalf("Failed to get entry from database: %v", err)
    }

    fmt.Printf("GetEntryFromDB: %s\n", e)
    return e
}

func (e *Entry) UpdateLink(db *sql.DB, newLink string) {
    e.Link = newLink

    // Prepate update statement
    statement, err := db.Prepare("UPDATE entries SET link = ? WHERE userID = ? and title = ? and category = ?")
    if err != nil {
        log.Fatalf("Failed to prepare statement: %v", err)
    }
    defer statement.Close()

    // Execute update statement
    _, err = statement.Exec(e.Link, e.userID, e.Title, e.Category)
    if err != nil {
        log.Fatalf("Failed to execute statement: %v", err)
    }

    fmt.Printf("Updated link for %s\n", e)
}


// Class methods
func (c Category) IsValid() error {
    switch c {
        case Movie, Show, Anime:
            return nil
        default:
            return fmt.Errorf("invalid category: %s", c)
    }
}

func (e *Entry) IsValid() error {

    if e.UserID == "" {
        return fmt.Errorf("user_id is empty")
    }

    if e.Title == "" {
        return fmt.Errorf("title is empty")
    }

    if err := e.Category.IsValid(); err != nil {
        return err
    }

    if date, err := time.Parse(TIME_FORMAT, e.Date); err != nil {
        return fmt.Errorf("invalid date: %s", e.Date)
    }
}

func (e *Entry) String() string {
    return fmt.Sprintf("UserID: %s\nTitle: %s\nCategory: %s\nDate: %s\nLink: %s\n",
                        e.UserID, e.Title, e.Category, e.Date, e.Link)
}
