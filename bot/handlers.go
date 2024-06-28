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
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
)

const (
	// Commands that the user will use to interact with the bot
	ENTRYPOINT      = "./watchlist"
	ADD_COMMAND     = "add"
	DELETE_COMMAND  = "delete"
	VIEW_COMMAND    = "view"
	UPDATE_COMMAND  = "update"
	HELP_COMMAND    = "help"
	CONTACT_COMMAND = "contact"

	// possible new features
	// LETTERBOXD_COMMAND 	= "letterboxd"	// random movie from letterboxd list
	// IMDB_COMMAND 		= "imdb"		// random movie from imdb list
	// MAL_COMMAND 			= "mal"			// random anime from myanimelist list
)

// Matches words and quoted strings (ex. Godfather, "The Godfather")
//
//	"[^"]+"     matches 1+ substrings inside quotes
//	\S+         matches 1+ substrings separated by whitespaces
var REGEX_PATTERN = regexp.MustCompile(`("[^"]+"|\S+)`)

// Main handler for the bot that will delegate to private handlers based on user input
func MasterHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from self
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Use regex to parse command and arguments from message
	args := REGEX_PATTERN.FindAllString(m.Content, -1)

	// Send help to messages without commands
	if len(args) < 2 {
		helpHandler(s, m)
	}

	// Ignore messages not addressed to us
	if args[0] != ENTRYPOINT {
		return
	}

	// Fire the correct handler based on given command
	switch args[1] {
	case ADD_COMMAND:
		addHandler(db, s, m)
	case DELETE_COMMAND:
		deleteHandler(db, s, m)
	case VIEW_COMMAND:
		viewHandler(db, s, m)
	case UPDATE_COMMAND:
		updateHandler(db, s, m)
	case HELP_COMMAND:
		helpHandler(s, m)
	case CONTACT_COMMAND:
		contactHandler(s, m)
	default:
		helpHandler(s, m)
	}
}

/* PRIVATE FUNCTIONS */

// Creates an entry and adds it to the watchlist, then sends a confirmation message
// Usage:
//
//	./watchlist add <title> <category> <link?>
//
// Example:
//
//	./watchlist add "The Godfather" movie
//	./watchlist add "The Godfather" movie "https://www.imdb.com/title/tt0133093/"
func addHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

	// Use regex to parse args from message
	args := REGEX_PATTERN.FindAllString(m.Content, -1)
	if len(args) < 4 {
		slog.Error("handlers.AddHandler", "msg", NotEnoughArgumentsError{m.Content})
		return
	} // Ensure we have at least a title and category

	// Validate and extract fields from args
	var title, link string
	var category Category

	if len(args) >= 4 {
		title = args[2]
		category = Category(args[3])
	}

	if len(args) == 5 {
		link = args[4]
	}

	// Create entry using validated fields
	entry := &Entry{
		UserID:   m.Author.ID,
		Title:    title,
		Category: category,
		Date:     time.Now(),
		Link:     link,
	}

	// Fetch watchlist from database
	watchlist, err := FetchWatchlist(db, m.Author.ID)
	if err != nil {
		slog.Error("handlers.AddHandler", "msg", err)
	}

	// Add entry to watchlist & database
	err = watchlist.Add(db, entry)
	if err != nil {
		slog.Error("handlers.AddHandler", "msg", err)
	}

	// Log and send a confirmation message
	slog.Info("handlers.AddHandler", "user", m.Author.Username, "entry", entry)
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```added %s to your watchlist```", entry.Title))
}

// Deletes an entry from the watchlist, then sends a confirmation message
// Usage:
//
//	./watchlist remove <title>
//	./watchlist remove <title> <category>
//
// Example:
//
//	./watchlist remove "The Godfather"
//	./watchlist remove "The Godfather" movie
func deleteHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

	// Use regex to parse args from message
	args := REGEX_PATTERN.FindAllString(m.Content, -1)
	if len(args) < 3 {
		return
	}

	title := args[2]
	var category Category
	if len(args) >= 4 {
		category = Category(args[3])
	}

	// TODO skip FETCH, just DELETE from database we don't need any structs at all
	watchlist, err := FetchWatchlist(db, m.Author.ID)
	if err != nil {
		slog.Error("handlers.DeleteHandler", "msg", err)
	}

	// Delete entry from watchlist
	err = watchlist.Delete(db, m.Author.ID, title, category)
	if err != nil {
		slog.Error("handlers.DeleteHandler", "msg", err)
	}

	// Log and send a confirmation message
	slog.Info("handlers.DeleteHandler", "user", m.Author.Username, "entry.Title", title)
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```deleted %s from your watchlist```", title))
}

// Updates an entry link in the watchlist, then sends a confirmation message
// Usage:
//
//	./watchlist update <title> <new_link>
//	./watchlist update <title> <category> <new_link>
//
// Example:
//
//	./watchlist update "The Godfather" https://www.youtube.com/watch?v=UaVTIH8mujA
//	./watchlist update "The Godfather" movie https://www.youtube.com/watch?v=UaVTIH8mujA
func updateHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

	// Use regex to parse args from message
	args := REGEX_PATTERN.FindAllString(m.Content, -1)
	if len(args) < 4 {
		slog.Error("handlers.UpdateHandler", "msg", NotEnoughArgumentsError{m.Content})
		return
	} // Ensure we have at least a title and category

	// Validate and extract fields from args
	var title, newLink string
	var category Category

	if len(args) == 4 {
		title = args[2]
		newLink = args[3]
	}

	if len(args) >= 5 {
		title = args[2]
		category = Category(args[3])
		newLink = args[4]
	}

	// TODO skip FETCH, just UPDATE from database we don't need any structs at all
	watchlist, err := FetchWatchlist(db, m.Author.ID)
	if err != nil {
		slog.Error("handlers.UpdateHandler", "msg", err)
	}

	// Update database
	err = watchlist.Update(db, m.Author.ID, title, category, newLink)
	if err != nil {
		slog.Error("handlers.UpdateHandler", "msg", err)
	}

	// Log and send a confirmation message
	slog.Info("handlers.UpdateHandler", "user", m.Author.Username, "entry.Title", title)
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```updated %s -> %s```", title, newLink))
}

// Displays the watchlist (sorted by position), then sends a confirmation message
// Usage:
//
//	./watchlist view <sort_by?>
//
// Example:
//
//	./watchlist view
//	./watchlist view title
//	./watchlist view date
//	./watchlist view category
func viewHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

	args := REGEX_PATTERN.FindAllString(m.Content, -1)
	var sort_by Sort_by
	if len(args) >= 3 {
		sort_by = Sort_by(args[2])
	}

	// Fetch & sort watchlists
	watchlist, err := FetchWatchlist(db, m.Author.ID)
	if err != nil {
		slog.Error("handlers.ViewHandler", "msg", err)
	}

	watchlist.Sort(sort_by)

	// Convert watchlist entries into a list of embed fields
	var embedFields []*discordgo.MessageEmbedField
	for _, entry := range watchlist.Entries {
		embedFields = append(embedFields, &discordgo.MessageEmbedField{
			Name:   entry.Title,
			Value:  fmt.Sprintf("(%s) %s)", entry.Category, entry.Link),
			Inline: true,
		})
	}

	// Create a thumbnail using the author's avatar
	thumbnail := &discordgo.MessageEmbedThumbnail{
		URL: m.Author.AvatarURL(""), // empty string for default avatar size
	}

	// Create and send embedded message
	embed := &discordgo.MessageEmbed{
		Fields:    embedFields,
		Thumbnail: thumbnail,
	}

	// Log and send watchlist
	slog.Info("handlers.ViewHandler", "user", m.Author.Username, "sort_by", sort_by, "watchlist", watchlist)
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// Displays the help message
// Usage:
//
//	./watchlist help
//	./watchlist help <command>
//
// Example:
//
//	./watchlist help
//	./watchlist help add
func helpHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// This will match all of the following cases
	//      ./watchlist help
	//      ./watchlist help <COMMAND>
	//      ./watchlist <COMMAND> help
	addMessage := "Adding a movie to your watchlist:\n```./watchlist add <title> <category> <position(optional)> <link(optional)>```\n"
	delMessage := "Deleting a movie from your watchlist:\n```./watchlist remove <title>```\n"
	viewMessage := "Viewing your watchlist:\n```./watchlist view\n./watchlist view title\n./watchlist view category\n./watchlist view date```\n"
	updateMessage := "Updating a movie in your watchlist:\n```./watchlist update <title> <new_link>```\n"
	helpMessage := "Displaying this help message:\n```./watchlist help\n./watchlist help <command>```\n"

	slog.Info("HelpHandler", "user", m.Author.Username)

	// If the user's help request is more specific, display the relevant help message'
	if strings.Contains(m.Content, ADD_COMMAND) {
		s.ChannelMessageSend(m.ChannelID, addMessage)
	} else if strings.Contains(m.Content, DELETE_COMMAND) {
		s.ChannelMessageSend(m.ChannelID, delMessage)
	} else if strings.Contains(m.Content, VIEW_COMMAND) {
		s.ChannelMessageSend(m.ChannelID, viewMessage)
	} else if strings.Contains(m.Content, UPDATE_COMMAND) {
		s.ChannelMessageSend(m.ChannelID, updateMessage)

		// If not, display all help messages
	} else {
		s.ChannelMessageSend(m.ChannelID, addMessage+delMessage+viewMessage+updateMessage+helpMessage)
	}
}

// Displays the contact message
// Usage:
//
//	./watchlist contact
func contactHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	slog.Info("handlers.ContactHandler", "user", m.Author.Username)
	s.ChannelMessageSend(m.ChannelID, "https://github.com/ttamre/go.watch")
}
