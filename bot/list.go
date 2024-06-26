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

    "github.com/bwmarrin/discordgo"
)

// Watchlist represents a user's watchlist
type Watchlist struct {
    UserID string
    Entries []Entry
}


// adds a new entry to the watchlist
func (w *Watchlist) Add(e Entry) {
    // Validate the entry
    if err := e.IsValid(); err != nil {
        log.Fatal("Failed to add entry: %v", err)
    }

    // Add the entry to the watchlist
    w.Entries = append(w.Entries, e)
}


// deletes an entry from the watchlist
func (w *Watchlist) Delete(e Entry) {
    // Find the entry to remove
    for i, watchlistItem := range w.Entries {
        if e == watchlistItem {
            // Remove the entry from the watchlist
            w.Entries = append(w.Entries[:i], w.Entries[i+1:]...)
        }
    }

    log.Fatalf("Failed to remove entry: %s not found", title)
}


// view the watchlist with an optional sorting parameter
func (w *Watchlist) View(sorting ...String) {

    switch sorting {

        // Sort by date
        case "date":
            sort.Slice(w.Entries, func(i, j int) bool {
                return w.Entries[i].Date < w.Entries[j].Date
            })

        // Sort by category
        case "category":
            sort.Slice(w.Entries, func(i, j int) bool {
                return w.Entries[i].Category < w.Entries[j].Category
            })

        // Sort by position
        default:

            sort.Slice(w.Entries, func(i, j int) bool {
                return w.Entries[i].Position < w.Entries[j].Position
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

    return watchlistString
}
