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


/*
Considerations

    structure
        a) single table     -> entries(userID, title, category, date, link)
        b) multiple tables  -> movies/shows/anime(userID, title, date, link)

        option a is more efficient
        option b is more scalable + easier to avoid duplicate titles of different categories
            (however, this problem can be solved with a composite key on (userID, title, and category))

    unique idenfiers
        a) primary key      -> itemID INTEGER PRIMARY KEY AUTOINCREMENT
        b) composite key    -> PRIMARY KEY (userID, title, category)

        option a is more efficient
        option b is safer for race conditions
*/
CREATE TABLE IF NOT EXISTS entries (
    userID      TEXT NOT NULL,
    title       TEXT NOT NULL,
    category    TEXT NOT NULL,
    date        DATETIME NOT NULL,
    link        TEXT,

    PRIMARY KEY (userID, title, category)
);
