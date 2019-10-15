package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Sets a reddit feed by subreddit and other params
func setRedditFeedCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		subreddit string
		author    string
		postType  = "hot"
		pin       bool
		title     string

		subIndex int
	)

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	cmdStrs := strings.Split(messageLowercase, " ")

	if len(cmdStrs) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"addfeed [u/author]* [type]* [pin]* [r/subreddit] [title]*`\n\n* are optional.\n\nType refers to the post sort filter. Valid values are `hot`, `new` and `rising`. Defaults to `hot`.\nPin refers to whether to pin the post when the bot posts it and unpin the previous bot pin of the same subreddit. Use `true` or `false` as values.\nTitle is what a post title should start with for the BOT to post it. Leave empty for all posts.\n\nFor author and subreddit be sure to add the prefixes `u/` and `r/`. Does not work with hidden or quarantined subs.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Finds where the subreddit and its index are and saves them
	for i, val := range cmdStrs {
		if strings.HasPrefix(val, "r/") || strings.HasPrefix(val, "/r/") {
			subreddit = strings.TrimPrefix(val, "/r/")
			subreddit = strings.TrimPrefix(subreddit, "r/")
			subIndex = i
			break
		}
	}

	if subreddit == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: subreddit not found. Please start it with `/r/` or `r/`.\n\nExample: `r/subreddit`.\n\nThis is not optional.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Now iterates all other parameters against the subreddit index which acts as a separator
	for i := subIndex; i > 0; i-- {

		// Saves pin true bool and its index if found
		if cmdStrs[i] == "true" || cmdStrs[i] == "1" || cmdStrs[i] == "positive" {
			pin = true
		}

		// Saves type and its index if found
		if cmdStrs[i] == "hot" || cmdStrs[i] == "rising" || cmdStrs[i] == "new" {
			postType = cmdStrs[i]
		}

		// Saves author if found with prefix
		if strings.HasPrefix(cmdStrs[i], "u/") || strings.HasPrefix(cmdStrs[i], "/u/") {
			author = strings.TrimPrefix(cmdStrs[i], "/u/")
			author = strings.TrimPrefix(author, "u/")
			break
		}
	}

	// Fetches title from after the subreddit index
	for i := subIndex + 1; i < len(cmdStrs); i++ {
		if i == len(cmdStrs)-1 {
			title += cmdStrs[i]
			break
		}
		title += cmdStrs[i] + " "
	}

	functionality.MapMutex.Lock()
	err := functionality.RssThreadsWrite(subreddit, author, title, postType, m.ChannelID, m.GuildID, pin)
	if err != nil {
		functionality.MapMutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	functionality.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! This reddit feed has been saved."))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a previously set reddit feed
func removeRedditFeedCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		subreddit string
		title     string
		postType  string
		author    string
		channelID string

		subIndex int
	)

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()

	// Check if there are set reddit feeds for this guild
	if functionality.GuildMap[m.GuildID].Feeds == nil || len(functionality.GuildMap[m.GuildID].Feeds) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error. There are no set reddit feeds.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}
	functionality.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	cmdStrs := strings.Split(messageLowercase, " ")

	if len(cmdStrs) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"removefeed [type]* [u/author]* [channel]* [r/subreddit] [title]*`\n\n* is optional\n\n"+
			"Type refers to the post sort filter. Valid values are `hot`, `new` and `rising`. Defaults to `hot`.\n"+
			"\nAuthor is the name of the post author.\n"+
			"\nChannel is the ID or name of a channel from which to remove\n"+
			"\nTitle is what a post title should start with or be for the BOT to post it. Leave empty for all feeds fulfilling [type] and [r/subreddit].")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fetches subreddit and its index
	for i, val := range cmdStrs {
		if strings.HasPrefix(val, "/r/") || strings.HasPrefix(val, "r/") {
			subreddit = strings.TrimPrefix(val, "/r/")
			subreddit = strings.TrimPrefix(subreddit, "r/")
			subIndex = i
			break
		}
	}

	if subreddit == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: subreddit not found. Please start it with `/r/` or `r/`.\n\nExample: `r/subreddit`.\n\nThis is not optional.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Saves type and author if found as well as channel
	for i, val := range cmdStrs {
		if i >= subIndex {
			continue
		}
		if strings.HasPrefix(val, "/u/") || strings.HasPrefix(val, "u/") {
			author = strings.TrimPrefix(cmdStrs[i], "/u/")
			author = strings.TrimPrefix(author, "u/")
		}
		if val == "hot" || val == "rising" || val == "new" {
			postType = val
		}
		chaID, _ := functionality.ChannelParser(s, val, m.GuildID)
		if chaID != "" {
			channelID = chaID
		}
	}

	// Fetches title from after the subreddit
	for i := subIndex + 1; i < len(cmdStrs); i++ {
		if i == len(cmdStrs)-1 {
			title += cmdStrs[i]
		} else {
			title += cmdStrs[i] + " "
		}
	}

	functionality.MapMutex.Lock()
	err := functionality.RssThreadsRemove(subreddit, title, author, postType, channelID, m.GuildID)
	if err != nil {
		functionality.MapMutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	functionality.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! This reddit feed has been removed."))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Prints all currently set reddit feeds
func viewRedditFeedCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		message      string
		splitMessage []string
	)

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()

	if len(functionality.GuildMap[m.GuildID].Feeds) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set reddit feeds.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}

	// Iterates through all the reddit feeds if they exist and adds them to the message string and print them
	for i := 0; i < len(functionality.GuildMap[m.GuildID].Feeds); i++ {
		// Format print string
		message += fmt.Sprintf("**r/%v**", functionality.GuildMap[m.GuildID].Feeds[i].Subreddit)
		if functionality.GuildMap[m.GuildID].Feeds[i].Author != "" {
			message += fmt.Sprintf(" - **u/%v**", functionality.GuildMap[m.GuildID].Feeds[i].Author)
		}
		message += fmt.Sprintf(" - **%v**", functionality.GuildMap[m.GuildID].Feeds[i].PostType)
		if functionality.GuildMap[m.GuildID].Feeds[i].Pin {
			message += " - **pinned**"
		}
		message += fmt.Sprintf(" - **%v**", functionality.GuildMap[m.GuildID].Feeds[i].ChannelID)
		if functionality.GuildMap[m.GuildID].Feeds[i].Title != "" {
			message += fmt.Sprintf(" - **%v**", functionality.GuildMap[m.GuildID].Feeds[i].Title)
		}
		message += "\n"
	}
	functionality.MapMutex.Unlock()

	// Splits the message if it's over 1900 characters
	if len(message) > 1900 {
		splitMessage = functionality.SplitLongMessage(message)
	}

	// Prints split or unsplit whois
	if splitMessage == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for i := 0; i < len(splitMessage); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot send feed message.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    setRedditFeedCommand,
		Trigger:    "addfeed",
		Aliases:    []string{"setfeed", "adfeed", "addreddit", "setreddit"},
		Desc:       "Adds a reddit feed to the channel",
		Permission: functionality.Mod,
		Module:     "reddit",
	})
	functionality.Add(&functionality.Command{
		Execute:    removeRedditFeedCommand,
		Trigger:    "removefeed",
		Aliases:    []string{"killfeed", "deletefeed", "removereddit", "killreddit", "deletereddit"},
		Desc:       "Removes a reddit feed",
		Permission: functionality.Mod,
		Module:     "reddit",
	})
	functionality.Add(&functionality.Command{
		Execute:    viewRedditFeedCommand,
		Trigger:    "feeds",
		Aliases:    []string{"showreddit", "redditview", "redditshow", "printfeed", "viewfeeds", "showfeeds", "showfeed", "viewfeed", "feed"},
		Desc:       "Prints all currently set Reddit feeds",
		Permission: functionality.Mod,
		Module:     "reddit",
	})
}
