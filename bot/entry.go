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


// Entry represents a single entry in the watchlist
type Entry struct {
    Title       string
    Category    Category
    Position    int
    Date        string
    Link        string
}

// Category represents the type of item in the watchlist
type Category string
const (
    Movie   Category = "movie"
    Show    Category = "show"
    Anime   Category = "anime"
)


// Update methods for entry struct
func (e *Entry) UpdateTitle(newTitle string) {
    e.Title = newTitle
}

func (e *Entry) UpdateCategory(newCategory Category) {
    e.Category = newCategory
}

func (e *Entry) UpdatePosition(newPosition int) {
    e.Position = newPosition
}

func (e *Entry) UpdateLink(newLink string) {
    e.Link = newLink
}


/* VALIDATION FUNCTONS */

func (c *Category) IsValid() error {
    switch c:
        case Movie, Show, Anime:
            return nil
        default:
            return fmt.Errorf("invalid category: %s", c)
}


func (e *Entry) IsValid() error {
    if e.Position < 0 {
        return fmt.Errorf("invalid position: %d", e.Position)
    }
    return e.Category.IsValid()
}
