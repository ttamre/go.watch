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

import "fmt"

/* STRUCTS */

type InvalidUserIDError struct {
	userID string
}

type InvalidTitleError struct {
	title string
}

type InvalidCategoryError struct {
	category *Category
}

type InvalidTimestampError struct {
	time string
}

type InvalidSortByError struct {
	sortBy *SortBy
}

type NotEnoughArgumentsError struct {
	message string
}

/* CLASS METHODS */

func (e *InvalidUserIDError) Error() string {
	return fmt.Sprintf("invalid user ID: %s", e.userID)
}

func (e *InvalidTitleError) Error() string {
	return fmt.Sprintf("invalid title: %s", e.title)
}

func (e *InvalidCategoryError) Error() string {
	return fmt.Sprintf("invalid category option: %s", *e.category)
}

func (e *InvalidTimestampError) Error() string {
	return fmt.Sprintf("invalid timestamp: %s", e.time)
}

func (e *InvalidSortByError) Error() string {
	return fmt.Sprintf("invalid sorting option: %s", *e.sortBy)
}

func (e *NotEnoughArgumentsError) Error() string {
	return fmt.Sprintf("not enough arguments: %s", e.message)
}
