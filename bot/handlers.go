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
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
)

const (
	// Commands that the user will use to interact with the bot
	ENTRYPOINT = "./watchlist"

	ADD_COMMAND     = "add"     // Add entry to watchlist
	DELETE_COMMAND  = "delete"  // Delete item from watchlist
	VIEW_COMMAND    = "view"    // View watchlist
	UPDATE_COMMAND  = "update"  // Update the link for an entry
	DONE_COMMAND    = "done"    // Mark entry as complete
	RATE_COMMAND    = "rate"    // Rate an entry
	RANDOM_COMMAND  = "random"  // Get a random movie from watchlist
	CONTACT_COMMAND = "contact" // Get contact info for the developer
	HELP_COMMAND    = "help"    // Display help message

	// LETTERBOXD_COMMAND 	= "letterboxd"	// random movie from letterboxd list
	// IMDB_COMMAND 		= "imdb"		// random movie from imdb list
	// MAL_COMMAND 			= "mal"			// random anime from myanimelist list
)

// Matches words and quoted strings (ex. Godfather, "The Godfather")
//
//	"[^"]+"     matches 1+ substrings inside quotes
//	\S+         matches 1+ substrings separated by whitespaces
var REGEX_PATTERN = regexp.MustCompile(`("[^"]+"|\S+)`)

/*
Main handler for the bot that will delegate to private handlers based on user input

All handler functions require use the following parameters:
  - db: ptr to database connection (not required for help and contact handlers)
  - s:  ptr to discord session (contains methods for websocket communication)
  - m:  ptr to discord message (contains info about author, channel, etc.)
*/
func MasterHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from self
	if m.Author.ID == s.State.User.ID {
		return
	}

	// args = []string{"./watchlist <command> <arg1> <arg2> ..."}
	args := REGEX_PATTERN.FindAllString(m.Content, -1)

	// Ignore messages not addressed to us
	if args[0] != ENTRYPOINT {
		return
	}

	// Send help to messages without commands
	if len(args) < 2 {
		helpHandler(s, m)
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
	case DONE_COMMAND:
		doneHandler(db, s, m)
	case RATE_COMMAND:
		rateHandler(db, s, m)
	case RANDOM_COMMAND:
		randomHandler(db, s, m)
	case HELP_COMMAND:
		helpHandler(s, m)
	case CONTACT_COMMAND:
		contactHandler(s, m)
	default:
		// if invalid command, send help message
		helpHandler(s, m)
	}
}

/*
Creates an entry and adds it to the watchlist, then sends a confirmation message

Usage:

	./watchlist add <title> <category> <link?>

Example:

	./watchlist add "The Godfather" movie
	./watchlist add "The Godfather" movie "https://www.imdb.com/title/tt0133093/"
*/
func addHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

	// args = []string{"./watchlist", add, title, category, link?}
	args := REGEX_PATTERN.FindAllString(m.Content, -1)
	if len(args) < 4 {
		slog.Error("handlers.addHandler", "msg", NotEnoughArgumentsError{m.Content})
		return
	} // Ensure we have at least a title and category

	title := args[2]
	category := Category(args[3])

	// case: ./watchlist add <movie> <category> <link>
	var link string
	if len(args) == 5 {
		link = args[4]
	}

	entry, err := NewEntry(m.Author.ID, title, category, link)
	if err != nil {
		slog.Error("handlers.AddHandler", "msg", err)
		return
	}

	// Add to database
	err = AddEntry(db, entry)
	if err != nil {
		slog.Error("handlers.AddHandler", "msg", err)
		return
	}

	// Log and send a confirmation message
	slog.Info("handlers.AddHandler", "user", m.Author.Username, "entry", entry)
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```added %s to your watchlist```", entry.Title))
}

/*
Deletes an entry from the watchlist, then sends a confirmation message

Usage:

	./watchlist delete <title>
	./watchlist delete <title> <category>

Example:

	./watchlist delete "The Godfather"
	./watchlist delete "The Godfather" movie
*/
func deleteHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

	// args = []string{"./watchlist", title, category?}
	args := REGEX_PATTERN.FindAllString(m.Content, -1)
	if len(args) < 3 {
		slog.Error("handlers.deleteHandler", "msg", NotEnoughArgumentsError{m.Content})
		return
	} // Ensure we have at least a title

	title := args[2]
	var category Category

	// case ./watchlist remove <title> <category>
	if len(args) >= 4 {
		category = Category(args[3])
	}

	// Delete entry
	err := DeleteEntry(db, m.Author.ID, title, category)
	if err != nil {
		slog.Error("handlers.DeleteHandler", "msg", err)
		return
	}

	// Log and send a confirmation message
	slog.Info("handlers.DeleteHandler",
		"user", m.Author.Username,
		"title", title,
		"category", category,
	)
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```deleted %s from your watchlist```", title))
}

/*
Displays the watchlist (sorted by position), then sends a confirmation message

Usage:

	./watchlist view <sort_by?>

Example:

	./watchlist view
	./watchlist view title
	./watchlist view date
	./watchlist view category
*/
func viewHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

	// args = []string{"./watchlist", "view", sort_by}
	args := REGEX_PATTERN.FindAllString(m.Content, -1)
	// no need to verify args because we have a default value for sort_by

	sort_by := SORT_WATCHED
	if len(args) >= 3 {
		sort_by = SortBy(args[2])
	}

	// Fetch watchlist (including watched items) & sort
	watchlist, err := FetchWatchlist(db, m.Author.ID, true)
	if err != nil {
		slog.Error("handlers.viewHandler", "msg", err)
		return
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

	// Create embedded message with entries and thumbnail
	embed := &discordgo.MessageEmbed{
		Fields:    embedFields,
		Thumbnail: thumbnail,
	}

	// Log and send watchlist as an embedded message
	slog.Info("handlers.viewHandler",
		"user", m.Author.Username,
		"sort_by", sort_by,
		"watchlist", watchlist)
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

/*
Updates an entry link in the watchlist, then sends a confirmation message
Usage:

	./watchlist update <title> <new_link>
	./watchlist update <title> <category> <new_link>

Example:

	./watchlist update "The Godfather" https://www.youtube.com/watch?v=UaVTIH8mujA
	./watchlist update "The Godfather" movie https://www.youtube.com/watch?v=UaVTIH8mujA
*/
func updateHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

	// args1 = []string{"./watchlist", update, title, category}
	// args2 = []string{"./watchlist", update, title, category, new_link}
	args := REGEX_PATTERN.FindAllString(m.Content, -1)
	if len(args) < 4 {
		slog.Error("handlers.updateHandler", "msg", NotEnoughArgumentsError{m.Content})
		return
	} // Ensure we have at least a title and category

	var (
		title, newLink string
		category       Category
	)

	// case 1: ./watchlist update <title> <new_link>
	if len(args) == 4 {
		title = args[2]
		newLink = args[3]
	}

	// case 2: ./watchlist update <title> <category> <new_link>
	if len(args) >= 5 {
		title = args[2]
		category = Category(args[3])
		newLink = args[4]
	}

	// Update database
	err := UpdateEntry(db, m.Author.ID, title, category, newLink)
	if err != nil {
		slog.Error("handlers.updateHandler", "msg", err)
		return
	}

	// Log and send a confirmation message
	slog.Info("handlers.updateHandler", "user", m.Author.Username, "title", title)
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```updated %s -> %s```", title, newLink))
}

/*
Marks an entry as complete, then sends a confirmation message
Usage:

	./watchlist done <title>
	./watchlist done <title> <category>

Example:

	./watchlist done "The Godfather"
	./watchlist done "The Godfather" movie
*/
func doneHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

	// args = []string{"./watchlist", "done", title, category}
	args := REGEX_PATTERN.FindAllString(m.Content, -1)
	if len(args) < 4 {
		slog.Error("handlers.doneHandler", "msg", NotEnoughArgumentsError{m.Content})
		return
	} // Ensure we have at least a title and category

	var (
		title    string
		category Category
	)

	// case: ./watchlist done <title>
	// case: ./watchlist update <title> <category> <new_link>
	if len(args) >= 3 {
		title = args[2]
	}

	// case: ./watchlist done <title> <category>
	if len(args) >= 4 {
		category = Category(args[3])
	}

	// Update database
	err := DoneEntry(db, m.Author.ID, title, category)
	if err != nil {
		slog.Error("handlers.doneHandler", "msg", err)
		return
	}

	// Log and send a confirmation message
	slog.Info("handlers.doneHandler", "user", m.Author.Username, "title", title)
	message := fmt.Sprintf("```completed %s\nrate it with ./watchlist %s %s <rating>```", title, RATE_COMMAND, title)
	s.ChannelMessageSend(m.ChannelID, message)
}

/*
Marks an entry as complete, then sends a confirmation message

Usage:

	./watchlist rate <title> <rating>
	./watchlist rate <title> <category> <rating>

Example:

	./watchlist rate "The Godfather" 5
	./watchlist rate "The Godfather" movie 5
*/
func rateHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

	// args1 = []string{"./watchlist", "rate", title, rating}
	// args1 = []string{"./watchlist", "rate", title, category, rating}
	args := REGEX_PATTERN.FindAllString(m.Content, -1)
	if len(args) < 4 {
		slog.Error("handlers.rateHandler", "msg", NotEnoughArgumentsError{m.Content})
		return
	} // Ensure we have at least a title and category

	// Validate and extract fields from args
	var (
		title    string
		rating   int
		category Category

		// Unconventional way of initializing an error,
		// but := cannot assign a value to err in the same line as rating
		// and rating needs to be pre-declared
		err error
	)

	// case 1: ./watchlist rate <title> <rating>
	if len(args) == 4 {
		title = args[2]

		rating, err = strconv.Atoi(args[3])
		if err != nil {
			slog.Error("handlers.rateHandler", "msg", err)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```invalid rating for %s: %s```", title, args[3]))
			return
		}
	}

	// case 2: ./watchlist rate <title> <category> <rating>
	if len(args) >= 5 {
		title = args[2]
		category = Category(args[3])

		rating, err = strconv.Atoi(args[4])
		if err != nil {
			slog.Error("handlers.rateHandler", "msg", err)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```invalid rating for %s: %s```", title, args[4]))
			return
		}
	}

	// Update database
	err = RateEntry(db, m.Author.ID, title, category, rating)
	if err != nil {
		slog.Error("handlers.rateHandler", "msg", err)
		return
	}

	// Log and send a confirmation message
	slog.Info("handlers.rateHandler",
		"user", m.Author.Username,
		"title", title,
		"rating", rating,
	)
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```rated %s %d stars```", title, rating))
}

/*
Gets a random movie from user's watchlist, then sends a confirmation message

Usage:

	./watchlist random
*/
func randomHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

	// Fetch watchlist (excluding watched entries)
	unwatched, err := FetchWatchlist(db, m.Author.ID, false)
	if err != nil {
		slog.Error("handlers.randomHandler", "msg", err)
		return
	}

	idx := rand.Intn(len(unwatched.Entries) - 1)
	entry := unwatched.Entries[idx]

	// Create a thumbnail using the author's avatar
	thumbnail := &discordgo.MessageEmbedThumbnail{
		URL: m.Author.AvatarURL(""), // empty string for default avatar size
	}

	// Create embedded message with entries and thumbnail
	embed := &discordgo.MessageEmbed{
		Title:     entry.Title,
		URL:       entry.Link,
		Thumbnail: thumbnail,
		Timestamp: entry.Date.Format(time.RFC3339),
	}

	// Log and send watchlist as an embedded message
	slog.Info("handlers.randomHandler", "user", m.Author.Username, "unwatched", unwatched)
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

/*
Displays the help message

Usage:

	./watchlist help
	./watchlist help <command>

Example:

	./watchlist help
	./watchlist help add
*/
func helpHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// args = []string{"./watchlist", "help", command}
	args := REGEX_PATTERN.FindAllString(m.Content, -1)

	var command string
	if len(args) >= 3 {
		command = args[2]
	}

	addMessage := "Adding a movie to your watchlist:\n```./watchlist add <title> <category> <position(optional)> <link(optional)>```"
	delMessage := "Deleting a movie from your watchlist:\n```./watchlist remove <title>```"
	viewMessage := "Viewing your watchlist:\n```./watchlist view\n./watchlist view title\n./watchlist view category\n./watchlist view date```"
	updateMessage := "Updating a movie in your watchlist:\n```./watchlist update <title> <new_link>```"
	doneMessage := "Marking a movie as completed:\n```./watchlist done <title>\n./watchlist done <title> <category>```"
	rateMessage := "Rating a movie in your watchlist:\n```./watchlist rate <title> <rating>\n./watchlist rate <title> <category> <rating>```"
	randomMessage := "Getting a random movie from your watchlist:\n```./watchlist random```"
	helpMessage := "Displaying this help message:\n```./watchlist help\n./watchlist help <command>```"
	contactMessage := "Get contact info for the developer:\n```./watchlist contact```"

	messages := map[string]string{
		ADD_COMMAND:     addMessage,
		DELETE_COMMAND:  delMessage,
		VIEW_COMMAND:    viewMessage,
		UPDATE_COMMAND:  updateMessage,
		DONE_COMMAND:    doneMessage,
		RATE_COMMAND:    rateMessage,
		RANDOM_COMMAND:  randomMessage,
		HELP_COMMAND:    helpMessage,
		CONTACT_COMMAND: contactMessage,
	}

	message, ok := messages[command]

	// If we get no command or an invalid command, show all help tips
	if !ok {
		for _, message := range messages {
			message += "\n"
		}
	}

	slog.Info("handlers.HelpHandler", "user", m.Author.Username)
	s.ChannelMessageSend(m.ChannelID, message)
}

/*
Displays the contact message

	Usage:

		./watchlist contact
*/
func contactHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	slog.Info("handlers.ContactHandler", "user", m.Author.Username)
	s.ChannelMessageSend(m.ChannelID, "https://github.com/ttamre/watchlist")
}
