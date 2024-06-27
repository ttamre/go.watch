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
    "log"
    "sort"
    "strings"
)

/* STRUCTS, ENUMS, CONSTANTS */
type Watchlist struct {
    UserID string   `json:"user_id"`
    Entries []Entry `json:"entries"`
}

type Sort_by string
const (
    // Enumerations for sorting the watchlist with the view command
    sort_by_title       Sort_by = "title"
    sort_by_date        Sort_by = "date"
    sort_by_category    Sort_by = "category"
)


/* CUSTOM ERROR TYPES */
type WatchlistDoesntExistError struct {
    message string
}

func (e *WatchlistDoesntExistError) Error() string {
    return e.message
}


/* FUNCTIONS */

// Fetch the watchlist from the database if exists, creates a new one otherwise
func FetchWatchlist(db *sql.DB, userID string) *Watchlist {
    // Create an empty watchlist
    watchlist := &Watchlist{UserID: userID}

    // If an entry for the user exists, get it + all other entries
    if checkWatchlist(db, userID) {
        watchlist.getEntries(db, userID)
    }

    return watchlist
}

func (w *Watchlist) getEntries(db *sql.DB, userID string) *Watchlist {
    // Get all entries from the database for the user
    query := "SELECT * FROM entries WHERE userID = ?"
    rows, err := db.Query(query, userID)
    if err != nil {
        log.Fatalf("Failed to get entries from database: %v", err)
    }

    defer rows.Close()

    var entries []Entry

    // Loop through row of query results and create Entry objects for each
    for rows.Next() {
        var e Entry
        err := rows.Scan(&e.UserID, &e.Title, &e.Category, &e.Date, &e.Link)
        if err != nil {
            log.Fatalf("Failed to scan entry from database: %v", err)
        }

        entries = append(entries, e)
    }

    w.Entries = entries
}

// Check if the watchlist exists in the database
func checkWatchlist(db *sql.DB, userID string) bool {
    var exists bool
    query := "SELECT EXISTS(SELECT 1 FROM entries WHERE userID = ? LIMIT 1)"
    err := db.QueryRow(query, userID).Scan(&exists)
    if err != nil {
        log.Fatalf("Failed to check if watchlist exists: %v", err)
    }

    return exists
}


/* CLASS METHODS */

// get entry from watchlist
func (w *Watchlist) Get(title string) *Entry {
    // sql -> SELECT * FROM entries WHERE userID = w.UserID AND title = title
    for _, entry := range w.Entries {
        if entry.Title == title {
            return &entry
        }
    }
}

// adds a new entry to the watchlist
func (w *Watchlist) Add(e Entry) {
    // Validate the entry
    if err := e.IsValid(); err != nil {
        log.Fatal("Failed to add entry: %v", err)
    }

    // Add the entry to the watchlist
    w.Entries = append(w.Entries, e)

    // Add the entry to the database
    // sql -> INSERT INTO entries (userID, title, category, date, link)
    //          VALUES e.UserID, e.Title, e.Category, e.Date, e.Link

    fmt.Println(w)
}

// deletes an entry from the watchlist
func (w *Watchlist) Delete(e Entry) {
    // Find the entry to remove
    for i, watchlistItem := range w.Entries {
        if e == watchlistItem {
            // Remove the entry from the watchlist
            w.Entries = append(w.Entries[:i], w.Entries[i+1:]...)
            // sql -> DELETE FROM entries
            // WHERE userID = e.UserID AND title = e.Title AND category = e.Category
            fmt.Println(w)
        }
    }
}

// Updates an entry's link in the watchlist
func (w *Watchlist) Update(entry *Entry, newLink string) {
    // Find the entry to update
    for i, oldEntry := range w.Entries {
        if newEntry == oldEntry {
            w.Entries[i].UpdateLink(newLink)
            // sql -> UPDATE entries SET link = newLink
            // WHERE userID = e.UserID AND title = e.Title AND e.Category = e.Category
            fmt.Println(w)
        }
    }
}


// view the watchlist (sorted by sort_by)
func (w *Watchlist) Sort(sort_by Sort_by) {

    // Validate the sort_by option
    // the list will be unsorted if the option is invalid, so this is redundant
    // but it's good practice to validate user input
    if err := sort_by.IsValid(); err != nil {
        log.War("Failed to sort watchlist: %v", err)
    }

    // no default case means list will not be sorted if an invalid enum is provided
    switch sort_by {

        case sort_by_title:
            sort.Slice(w.Entries, func(i, j int) bool {
                return w.Entries[i].Title < w.Entries[j].Title
            })

        case sort_by_date:
            sort.Slice(w.Entries, func(i, j int) bool {
                return time.Parse(TIME_FORMAT, w.Entries[i].Date)
                < time.Parse(TIME_FORMAT, w.Entries[j].Date)
            })

        case sort_by_category:
            sort.Slice(w.Entries, func(i, j int) bool {
                return w.Entries[i].Category < w.Entries[j].Category
            })
    }

    fmt.Println(w)
}

// string representation of a watchlist
func (w *Watchlist) String() string {
    var watchlistString = fmt.Sprintf("Watchlist for %s:\n", w.UserID)
    watchlistString = strings.Repeat("-", len(watchlistString)) // dotted line under the title

    for _, e := range w.Entries {
        watchlistString += fmt.Sprintf("\t(%d): %s (%s)\n", e.Position, e.Title, e.Category)
    }
    return watchlistString
}

// validation method for sort_by options
func (s *Sort_by) IsValid() error {
    switch *s {
        case sort_by_title, sort_by_date, sort_by_category:
            return nil
        default:
            return fmt.Errorf("invalid sort_by option: %s", *s)
    }
}
