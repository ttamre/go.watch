/*
./watchlist add     <title> <category> <position?> <link?>
./watchlist remove  <title>
./watchlist view    [date?, category?]
*/

package bot

import (
    "github.com/bwmarrin/discordgo"
    "github.com/go-redis/redis/v9"
)
