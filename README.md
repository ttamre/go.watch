<h1 style="font-family:monospace">watchlist</h1>
<div style="padding-bottom:20px">
    <img src="https://img.shields.io/badge/go-1.22.0-blue" />
    <img src="https://img.shields.io/badge/sqlite-3.32.3-grey" />
    <img src="https://img.shields.io/badge/license-GPL%20v3-green" />
</div>

<!-- DESCRIPTION -->
<p style="font-family:monospace">A watchlist manager discord bot, written in go + sqlite3<br></p>

<!-- INSTALLATION -->
<h2 style="font-family:monospace">Installation</h2>

```bash
# clone repo
git clone https://github.com/ttamre/watchlist.git
cd watchlist

# set this environment variable to your discord bot token
export DISCORD_WATCHLIST_BOT_TOKEN=""

# build binaries
make
```

<!-- USAGE -->
<h2 style="font-family:monospace">Usage</h2>

```bash
# run binary
./bin/watchlist

# help message (type this in a text channel that the bot has access to)
./watchlist help
```


<!-- COMMANDS -->
<h2 style="font-family:monospace">Commands</h2>


<h4 style="font-family:monospace">Add an entry to your watchlist</h4>

`./watchlist add <title> <category> <link?>`

| PARAMETER | TYPE | DESCRIPTION | REQUIRED |
| --------- | ---- | ----------- | -------- |
| title | `text` | title of the movie |✅|
| category | `text` | one of (movie/show/anime) | ✅|
| link | `text` | link to a trailer/imdb/etc |❌|


<h4 style="font-family:monospace">Delete an entry from your watchlist</h4>

`./watchlist delete <title> <category>`

| PARAMETER | TYPE | DESCRIPTION | REQUIRED |
| --------- | ---- | ----------- | -------- |
| title | `text` | title of the movie | ✅|
| category | `text` | one of (movie/show/anime) |❌|


<h4 style="font-family:monospace">View your watchlist</h4>

`./watchlist view <sorting>`

| PARAMETER | TYPE | DESCRIPTION | REQUIRED |
| --------- | ---- | ----------- | -------- |
| sorting | `text` | one of (date/title/category/watched/rating) |❌|


<h4 style="font-family:monospace">Update the link for an entry</h4>

`./watchlist update <title> <link>` or `./watchlist update <title> <category> <link>`

| PARAMETER | TYPE | DESCRIPTION | REQUIRED |
| --------- | ---- | ----------- | -------- |
| title | `text` | title of the movie | ✅|
| category | `text` | one of (movie/show/anime) |❌|
| newLink | `text` | link to a trailer/imdb/etc |✅|


<h4 style="font-family:monospace">Mark an entry as done</h4>

`./watchlist done <title> <category>`

| PARAMETER | TYPE | DESCRIPTION | REQUIRED |
| --------- | ---- | ----------- | -------- |
| title | `text` | title of the movie | ✅|
| category | `text` | one of (movie/show/anime) |❌|


<h4 style="font-family:monospace">Rate an entry</h4>

`./watchlist rate <title> <rating>>` or `./watchlist rate <title> <category> <rating>`

| PARAMETER | TYPE | DESCRIPTION | REQUIRED |
| --------- | ---- | ----------- | -------- |
| title | `text` | title of the movie | ✅|
| category | `text` | one of (movie/show/anime) |❌|
| rating | `int` | rating of the movie |✅|


<h4 style="font-family:monospace">Get a random entry from your watchlist</h4>

`./watchlist random`


<h4 style="font-family:monospace">Help message</h4>

`./watchlist help <command>`


| PARAMETER | TYPE | DESCRIPTION | REQUIRED |
| --------- | ---- | ----------- | -------- |
| command | `text` | command you need help with|❌|


<h4 style="font-family:monospace">Get the develper's contact info + source code</h4>

`./watchlist contact`



<!-- LICENSE -->
<h2 style="font-family:monospace">License</h2>
<p style="font-family:monospace">This project is licensed under the GNU v3 General Public License. For more information, see the <a href="LICENSE">LICENSE</a></p>
