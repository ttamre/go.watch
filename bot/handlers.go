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
    "strings"
    "time"
    "regexp"

    "github.com/bwmarrin/discordgo"
    "github.com/mattn/go-sqlite3"
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

    // Matches words and quoted strings (ex. Godfather, "The Godfather")
    //      "[^"]+"     matches 1+ substrings inside quotes
    //      \S+         matches 1+ substrings separated by whitespaces
    REGEX_PATTERN   = regexp.MustCompile(`("[^"]+"|\S+)`)
)

// Main handler for the bot that will delegate to private handlers based on user input
func MasterHandler(db *sql.DB, session *discordgo.Session, message *discordgo.MessageCreate) {
    // Ignore messages from self
    if message.Author.ID == session.State.User.ID { return }

    // Use regex to parse command and arguments from message
    args := REGEX_PATTERN.FindAllString(m.Content, -1)
    if len(args) < 2 { return }         // Ignore messages without commands
    if args[0] != ENTRYPOINT { return } // Ignore messages not addressed to us

    // Fire the correct handler based on given command
    switch args[1] {
        case ADD_COMMAND:
            addHandler(db, session, message)
        case DELETE_COMMAND:
            deleteHandler(db, session, message)
        case VIEW_COMMAND:
            viewHandler(db, session, message)
        case UPDATE_COMMAND:
            updateHandler(db, session, message)
        case HELP_COMMAND:
            helpHandler(session, message)
        case CONTACT_COMMAND:
            contactHandler(session, message)
        default:
            helpHandler(session, message)
    }
}


/* PRIVATE FUNCTIONS */

// Creates an entry and adds it to the watchlist, then sends a confirmation message
// Usage:
//      ./watchlist add <title> <category> <link?>
// Example:
//      ./watchlist add "The Godfather" movie
//      ./watchlist add "The Godfather" movie "https://www.imdb.com/title/tt0133093/"
func addHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

    // Use regex to parse args from message
    args := REGEX_PATTERN.FindAllString(m.Content, -1)
    if len(args) < 4 { return } // Ensure we have at least a title and category

    // Validate and extract fields from args
    var title, category, link string
    if len(args) >= 4 {
        title = args[2]
        category = Category(args[3])
    }

    if len(args) == 5 {
        link = args[4]
    }

    // Create entry using validated fields
    entry := Entry{
        UserID:     m.Author.ID,
        Title:      title,
        Category:   category,
        Date:       time.Now(),
        Link:       link,
    }

    // Fetch watchlist and add entry to it
    watchlist := FetchWatchlist(db, m.Author.ID)
    watchlist.Add(entry)

    // Send a confirmation message
    s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```AddHandler: %s```", watchlist))
}

// Deletes an entry from the watchlist, then sends a confirmation message
// Usage:
//     ./watchlist remove <title>
// Example:
//      ./watchlist remove "The Godfather"
func deleteHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

    // Use regex to parse args from message
    args := REGEX_PATTERN.FindAllString(m.Content, -1)
    if len(args) < 3 { return }

    // Get watchlist and search for entry to delete based on title
    watchlist := FetchWatchlist(db, m.Author.ID)
    entry := watchlist.Get(args[2])
    watchlist.Delete(entry)

    // Send a confirmation message
    s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```DeleteHandler: %s```", watchlist))
}

// Displays the watchlist (sorted by position), then sends a confirmation message
// Usage:
//      ./watchlist view <sort_by?>
// Example:
//      ./watchlist view
//      ./watchlist view title
//      ./watchlist view date
//      ./watchlist view category
func viewHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

    // use defined regex pattern to parse sort_by from message
    sort_string := "date"
    var sort_by Sort_by

    switch sort_string {
        case "title":
            sort_by = sort_by_title
        case "date":
            sort_by = sort_by_date
        case "category":
            sort_by = sort_by_category
        default:
            sort_by = sort_by_date
    }

    // Fetch & sort watchlists
    watchlist := FetchWatchlist(m.Author.ID)
    watchlist.Sort(sort_by)

    // Convert watchlist entries into a list of embed fields
    var embedFields []*discordgo.MessageEmbedField
    for _, entry := range watchlist.Entries {
        value := entry.Category.String())
        embedFields = append(embedFields, &discordgo.MessageEmbedField{
            Name: entry.Title,
            Value: fmt.Sprintf("(%s) %s)", entry.Category.String(), entry.Link),
            Inline: true,
        })
    }

    // Create a thumbnail using the author's avatar
    thumbnail := &discordgo.MessageEmbedThumbnail{
        URL: m.Author.AvatarURL(""), // empty string for default avatar size
    }

    // Create and send embedded message
    embed := &discordgo.MessageEmbed{
        Fields: embedFields,
        Thumbnail: thumbnail,
    }
    s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// Updates an entry link in the watchlist, then sends a confirmation message
// Usage:
//      ./watchlist update <title> <new_link>
// Example:
//      ./watchlist update "The Godfather" "https://www.youtube.com/watch?v=UaVTIH8mujA"
func updateHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {
    title := ""     // use regex to match titles in quotes with spaces
    newLink := ""   // use regex to get new link

    watchlist := FetchWatchlist(m.Author.ID)
    entry := watchlist.Get(title)
    watchlist.Update(entry, newLink)
    s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```ViewHandler: %s```", watchlist))
}

// Displays the help message
// Usage:
//      ./watchlist help
//      ./watchlist help <command>
// Example:
//      ./watchlist help
//      ./watchlist help add
func helpHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {

    // This will match all of the following cases
    //      ./watchlist help
    //      ./watchlist help <COMMAND>
    //      ./watchlist <COMMAND> help
    addMessage      := "Adding a movie to your watchlist:\n```./watchlist add <title> <category> <position(optional)> <link(optional)>```\n"
    delMessage      := "Deleting a movie from your watchlist:\n```./watchlist remove <title>```\n"
    viewMessage     := "Viewing your watchlist:\n```./watchlist view\n./watchlist view title\n./watchlist view category\n./watchlist view date```\n"
    updateMessage   := "Updating a movie in your watchlist:\n```./watchlist update <title> <new_link>```\n"
    helpMessage     := "Displaying this help message:\n```./watchlist help\n./watchlist help <command>```\n"

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
        s.ChannelMessageSend(m.ChannelID, addMessage + delMessage + viewMessage + updateMessage + helpMessage)
    }
}

// Displays the contact message
// Usage:
//      ./watchlist contact
func contactHandler(s *discordgo.Session, m *disordgo.MessageCreate) {
    s.ChannelMessageSend(m.ChannelID, "```For contact info, visit https://github.com/ttamre/go.watch```")
}
