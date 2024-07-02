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
	"reflect"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Entry represents a single entry in the watchlist
type Entry struct {
	UserID   string    `json:"user_id"`
	Date     time.Time `json:"date"`
	Title    string    `json:"title"`
	Category Category  `json:"category"`
	Done     bool      `json:"done"`
	Rating   int       `json:"rating"`
	Link     string    `json:"link"`
}

// Category represents the type of item in the watchlist
type Category string

const (
	Movie Category = "movie"
	Show  Category = "show"
	Anime Category = "anime"
)

/*
Create a new entry object with validation

Params:

	userID:		user ID of the entry
	title:		title of the entry
	category:	category of the entry
*/
func NewEntry(userID string, title string, category Category, link string) (*Entry, error) {
	e := &Entry{
		UserID:   userID,
		Date:     time.Now(),
		Title:    title,
		Category: category,
		Done:     false,
		Rating:   0,
		Link:     link,
	}

	// Validate entry
	if err := e.IsValid(); err != nil {
		return nil, err
	}

	return e, nil
}

/*
Adds an entry to the database

Params:

	db:	ptr to sqlite3 database connection
	e:	ptr to entry object
*/
func AddEntry(db *sql.DB, e *Entry) error {

	// Prepare insert statement
	query := "INSERT INTO entries(userID, date, title, category, done, rating, link) VALUES(?, ?, ?, ?, ?)"
	statement, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer statement.Close()

	// Execute insert statement
	_, err = statement.Exec(e.UserID, e.Date, e.Title, e.Category, e.Done, e.Rating, e.Link)
	if err != nil {
		return err
	}

	slog.Debug("entry.Add", "entry", e)
	return nil
}

/*
Delete an entry from the database

Params:

	db:		ptr to sqlite3 database connection
	userID:	user ID of the entry
*/
func DeleteEntry(db *sql.DB, userID string, title string, category Category) error {
	// Prepare delete statement
	statement, err := db.Prepare("DELETE FROM entries WHERE userID = ? and title = ? and category = ?")
	if err != nil {
		return err
	}
	defer statement.Close()

	// Execute delete statement
	_, err = statement.Exec(userID, title, category)
	if err != nil {
		return err
	}

	slog.Debug("entry.DeleteEntry", "user", userID, "title", title, "category", category)
	return nil
}

/*
Updates the link for an entry in the database

Params:

	db:			ptr to sqlite3 database connection
	userID:		user ID of the entry
	title:		title of the entry
	category:	category of the entry
	newLink:	new link to update the entry with
*/
func UpdateEntry(db *sql.DB, userID string, title string, category Category, newLink string) error {

	// Prepate update statement
	query := "UPDATE entries SET link = ? WHERE userID = ? and title = ? and category = ?"
	statement, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer statement.Close()

	// Execute update statement
	_, err = statement.Exec(newLink, userID, title, category)
	if err != nil {
		return err
	}

	slog.Debug("entry.UpdateEntry", "user", userID, "title", title, "category", category, "newLink", newLink)
	return nil
}

/*
Mark an entry as completed in the database

Params:

	db:			ptr to sqlite3 database connection
	userID:		user ID of the entry
	title:		title of the entry
	category:	category of the entry
*/
func DoneEntry(db *sql.DB, userID string, title string, category Category) error {
	// Prepare update statement
	query := "UPDATE entries SET done = 1 WHERE userID = ? and title = ? and category = ?"
	statement, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(userID, title, category)
	if err != nil {
		return err
	}

	slog.Debug("entry.DoneEntry", "user", userID, "title", title, "category", category)
	return nil
}

/*
Rate an entry in the database

Params:

	db:			ptr to sqlite3 database connection
	userID:		user ID of the entry
	title:		title of the entry
	category:	category of the entry
	rating:		rating to update the entry with
*/
func RateEntry(db *sql.DB, userID string, title string, category Category, rating int) error {
	// Prepare update statement
	query := "UPDATE entries SET rating = ? WHERE userID = ? and title = ? and category = ?"
	statement, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(rating, userID, title, category)
	if err != nil {
		return err
	}

	slog.Debug("entry.RateEntry", "user", userID, "title", title, "category", category, "rating", rating)
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
