/*
watchlist - a watchlist manager discord bot
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
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

/* STRUCTS, ENUMS, CONSTANTS */
type Watchlist struct {
	UserID  string   `json:"user_id"`
	Entries []*Entry `json:"entries"`
}

type Sort_by string

const (
	// Enumerations for sorting the watchlist with the view command
	sortTitle    Sort_by = "title"
	sortDate     Sort_by = "date"
	sortCategory Sort_by = "category"
)

/* FUNCTIONS */

// Fetch watchlist from the database if exists, creates a new one otherwise
func FetchWatchlist(db *sql.DB, userID string) (*Watchlist, error) {
	// Create an empty watchlist
	watchlist := &Watchlist{UserID: userID}

	// If an entry for the user exists, get it + all other entries
	exists, err := checkWatchlist(db, userID)

	if err != nil {
		return watchlist, err
	}

	if exists {
		err = watchlist.populate(db)
	}
	return watchlist, err
}

// Check if the watchlist exists in the database
func checkWatchlist(db *sql.DB, userID string) (bool, error) {
	exists := false
	query := "SELECT EXISTS(SELECT 1 FROM entries WHERE userID = ? LIMIT 1)"
	err := db.QueryRow(query, userID).Scan(&exists)
	return exists, err
}

/* CLASS METHODS */

// Populate a watchlist with entries that match the watchlist's user ID
func (w *Watchlist) populate(db *sql.DB) error {
	// Get all entries from the database for the user
	query := "SELECT (userID, title, category, date, link) FROM entries WHERE userID = ?"

	rows, err := db.Query(query, w.UserID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Loop through row of query results and create Entry objects for each
	var entries []*Entry
	for rows.Next() {
		var e Entry
		err := rows.Scan(&e.UserID, &e.Title, &e.Category, &e.Date, &e.Link)
		if err != nil {
			return err
		}

		entries = append(entries, &e)
	}

	w.Entries = entries
	return nil
}

// get entry from watchlist
func (w *Watchlist) Get(db *sql.DB, title string) (*Entry, error) {
	return GetEntryFromDB(db, w.UserID, title)

}

// adds a new entry to the watchlist and database
func (w *Watchlist) Add(db *sql.DB, e *Entry) error {
	// Validate the entry
	if err := e.IsValid(); err != nil {
		return err
	}

	// Add the entry to the watchlist
	w.Entries = append(w.Entries, e)

	// Add the entry to the database
	err := e.Add(db)
	if err != nil {
		return err
	}

	slog.Debug("watchlist.Add", "watchlist", w)
	return nil
}

// deletes an entry from the watchlist
func (w *Watchlist) Delete(db *sql.DB, userID string, title string, category Category) error {
	// Find the entry to remove
	for i, watchlistItem := range w.Entries {

		userIDMatch := userID == watchlistItem.UserID
		titleMatch := title == watchlistItem.Title
		categoryMatch := category == watchlistItem.Category

		if userIDMatch && titleMatch && categoryMatch {
			// Remove from watchlist
			w.Entries = append(w.Entries[:i], w.Entries[i+1:]...)

			// Remove from database
			err := watchlistItem.Delete(db)
			if err != nil {
				return err
			}

			slog.Debug("watchlist.Delete", "watchlist", w)
			return nil
		}
	}

	return &EntryNotFoundError{userID, title, category}
}

// Updates an entry's link in the watchlist
func (w *Watchlist) Update(db *sql.DB, userID string, title string, category Category, newLink string) error {
	// Find the entry to update
	for _, watchlistItem := range w.Entries {

		userIDMatch := userID == watchlistItem.UserID
		titleMatch := title == watchlistItem.Title
		categoryMatch := category == watchlistItem.Category

		if userIDMatch && titleMatch && categoryMatch {
			// Updates the entry + database
			err := watchlistItem.Update(db, newLink)
			if err != nil {
				return err
			}

			slog.Debug("watchlist.Update", "watchlist", w)
			return nil
		}
	}

	return &EntryNotFoundError{userID, title, category}
}

// view the watchlist (sorted by sort_by)
func (w *Watchlist) Sort(sort_by Sort_by) {

	// Validate the sort_by option (safe to proceed on invalid sort option)
	if err := sort_by.IsValid(); err != nil {
		slog.Warn("watchlist.Sort", "error", err)
	}

	// no default case --> list won't be sorted if invalid sort option is provided
	switch sort_by {

	case sortTitle:
		sort.Slice(w.Entries, func(i, j int) bool {
			return w.Entries[i].Title < w.Entries[j].Title
		})

	case sortDate:
		sort.Slice(w.Entries, func(i, j int) bool {
			return w.Entries[i].Date.Before(w.Entries[j].Date)
		})

	case sortCategory:
		sort.Slice(w.Entries, func(i, j int) bool {
			return w.Entries[i].Category < w.Entries[j].Category
		})
	}

	slog.Debug("watchlist.Sort", "watchlist", w)
}

// string representation of a watchlist
func (w *Watchlist) String() string {
	watchlistString := fmt.Sprintf("Watchlist for %s:\n", w.UserID)
	watchlistString += strings.Repeat("-", len(watchlistString))

	for _, e := range w.Entries {
		watchlistString += fmt.Sprintf("%s\n", e)
	}
	return watchlistString
}

// validation method for sort_by options
func (s *Sort_by) IsValid() error {
	switch *s {
	case sortTitle, sortDate, sortCategory:
		return nil
	default:
		return &InvalidSortByError{s}
	}
}
