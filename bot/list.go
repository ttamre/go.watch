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

type Watchlist struct {
	UserID  string   `json:"user_id"`
	Entries []*Entry `json:"entries"`
}

type SortBy string

const (
	// Enumerations for sorting the watchlist with the view command
	SORT_TITLE    SortBy = "title"
	SORT_DATE     SortBy = "date"
	SORT_CATEGORY SortBy = "category"
	SORT_WATCHED  SortBy = "watched"
	SORT_RATING   SortBy = "rating"
)

/*
Fetch watchlist from the database if it exists

Params:

	db: 		ptr to sqlite3 database connection
	userID: 	user ID we are searching for entries for
	watched:	true if we want all entries, false if we want only unwatched entries

Returns:

	*Watchlist: 	ptr to watchlist object
	error:			error object
*/
func FetchWatchlist(db *sql.DB, userID string, watched bool) (*Watchlist, error) {

	var watchlist *Watchlist

	// If an entry for the user exists, get it + all other entries
	exists, err := checkWatchlist(db, userID)
	if exists {
		watchlist := &Watchlist{UserID: userID}
		err = watchlist.populate(db, watched)
	}

	return watchlist, err
}

/*
Check if the watchlist exists in the database

Params:

	db: 		ptr to sqlite3 database connection
	userID: 	user ID we are searching for entries for

Returns:

	bool:		true if the watchlist exists, false otherwise
	error:		error object
*/
func checkWatchlist(db *sql.DB, userID string) (bool, error) {
	exists := false
	query := "SELECT EXISTS(SELECT 1 FROM entries WHERE userID = ? LIMIT 1)"
	err := db.QueryRow(query, userID).Scan(&exists)
	return exists, err
}

/*
Populate a watchlist with entries that match the watchlist's user ID

Params:

	db: 		ptr to sqlite3 database connection
	watched:	true if we want all entries, false if we want only unwatched entries

Returns:

	error:		error object
*/
func (w *Watchlist) populate(db *sql.DB, watched bool) error {
	// Get all entries from the database for the user
	query := "SELECT (userID, date, title, category, done, rating, link) " +
		"FROM entries WHERE userID = ?"

	if !watched {
		query += " AND done = 0"
	}

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

/*
Sort the watchlist by the provided sort_by option

Params:

	sort_by: 	sort_by option (one of SORT_TITLE, SORT_DATE, SORT_CATEGORY)
*/
func (w *Watchlist) Sort(sort_by SortBy) {

	// Validate the sort_by option (safe to proceed on invalid sort option)
	if err := sort_by.IsValid(); err != nil {
		slog.Warn("watchlist.Sort", "error", err)
	}

	// no default case --> list won't be sorted if invalid sort option is provided
	switch sort_by {

	case SORT_TITLE:
		sort.Slice(w.Entries, func(i, j int) bool {
			return w.Entries[i].Title < w.Entries[j].Title
		})

	case SORT_DATE:
		sort.Slice(w.Entries, func(i, j int) bool {
			return w.Entries[i].Date.Before(w.Entries[j].Date)
		})

	case SORT_CATEGORY:
		sort.Slice(w.Entries, func(i, j int) bool {
			return w.Entries[i].Category < w.Entries[j].Category
		})

	case SORT_WATCHED:
		sort.Slice(w.Entries, func(i, j int) bool {
			return w.Entries[i].Done == w.Entries[j].Done
		})

	case SORT_RATING:
		sort.Slice(w.Entries, func(i, j int) bool {
			return w.Entries[i].Rating < w.Entries[j].Rating
		})
	}

	slog.Debug("watchlist.Sort", "watchlist", w)
}

// stringer method
func (w *Watchlist) String() string {
	watchlistString := fmt.Sprintf("Watchlist for %s:\n", w.UserID)
	watchlistString += strings.Repeat("-", len(watchlistString))

	for _, e := range w.Entries {
		watchlistString += fmt.Sprintf("%s\n", e)
	}
	return watchlistString
}

// enum validation
func (s *SortBy) IsValid() error {
	switch *s {
	case SORT_TITLE, SORT_DATE, SORT_CATEGORY, SORT_WATCHED, SORT_RATING:
		return nil
	default:
		return &InvalidSortByError{s}
	}
}
