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
	"database/sql"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Entry represents a single entry in the watchlist
type Entry struct {
	UserID   string    `json:"user_id"`
	Title    string    `json:"title"`
	Category Category  `json:"category"`
	Date     time.Time `json:"date"`
	Link     string    `json:"link"`
}

// Category represents the type of item in the watchlist
type Category string

const (
	Movie Category = "movie"
	Show  Category = "show"
	Anime Category = "anime"
)

// Load the entry from the database
func GetEntryFromDB(db *sql.DB, userID string, title string) (*Entry, error) {

	var e Entry

	query := "SELECT (userID, title, category, date, link) " +
		"FROM entries WHERE userID = ? and title = ? LIMIT 1"

	err := db.QueryRow(query, userID, title).Scan(&e.UserID, &e.Title, &e.Category, &e.Date, &e.Link)
	if err != nil {
		return nil, err
	}

	slog.Debug("entry.GetEntryFromDB", "entry", e)
	return &e, nil
}

/* CLASS METHODS */

// Adds an entry to the database
func (e *Entry) Add(db *sql.DB) error {
	// Prepare insert statement
	query := "INSERT INTO entries(userID, title, category, date, link) VALUES(?, ?, ?, ?, ?)"
	statement, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer statement.Close()

	// Execute insert statement
	_, err = statement.Exec(e.UserID, e.Title, e.Category, e.Date, e.Link)
	if err != nil {
		return err
	}

	slog.Debug("entry.Add", "entry", e)
	return nil
}

// Deletes the entry from the database
func (e *Entry) Delete(db *sql.DB) error {
	// Prepare delete statement
	statement, err := db.Prepare("DELETE FROM entries WHERE userID = ? and title = ? and category = ?")
	if err != nil {
		return err
	}
	defer statement.Close()

	// Execute delete statement
	_, err = statement.Exec(e.UserID, e.Title, e.Category)
	if err != nil {
		return err
	}

	slog.Debug("entry.Delete", "entry", e)
	return nil
}

// Updates the link for an entry and applies changes to database
func (e *Entry) Update(db *sql.DB, newLink string) error {
	e.Link = newLink

	// Prepate update statement
	query := "UPDATE entries SET link = ? WHERE userID = ? and title = ? and category = ?"
	statement, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer statement.Close()

	// Execute update statement
	_, err = statement.Exec(e.Link, e.UserID, e.Title, e.Category)
	if err != nil {
		return err
	}

	slog.Debug("entry.Update", "entry", e)
	return nil
}

// Validator for category struct
func (c *Category) IsValid() error {
	switch *c {
	case Movie, Show, Anime:
		return nil
	default:
		return &InvalidCategoryError{c}
	}
}

// Validator for entry struct
func (e *Entry) IsValid() error {

	if e.UserID == "" {
		return &InvalidUserIDError{e.UserID}
	}

	if e.Title == "" {
		return &InvalidTitleError{e.Title}
	}

	if err := e.Category.IsValid(); err != nil {
		return err
	}

	if reflect.TypeOf(e.Date).String() != "time.Time" {
		return &InvalidTimestampError{e.Date.String()}
	}

	return nil
}

// Stringer for entry struct
func (e *Entry) String() string {
	if e.Link != "" {
		return fmt.Sprintf("%s (%s)\n%s\n", e.Title, e.Category, e.Link)
	} else {
		return fmt.Sprintf("%s (%s)\n", e.Title, e.Category)
	}
}
