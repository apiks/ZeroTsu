package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
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

	guildSettings := db.GetGuildSettings(m.GuildID)
	cmdStrs := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	if len(cmdStrs) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"addfeed [u/author]* [type]* [pin]* [r/subreddit] [title]*`\n\n* are optional.\n\nType refers to the post sort filter. Valid values are `hot`, `new` and `rising`. Defaults to `hot`.\nPin refers to whether to pin the post when the bot posts it and unpin the previous bot pin of the same subreddit. Use `true` or `false` as values.\nTitle is what a post title should start with for the BOT to post it. Leave empty for all posts.\n\nFor author and subreddit be sure to add the prefixes `u/` and `r/`. Does not work with hidden or quarantined subs.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
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
			common.LogError(s, guildSettings.BotLog, err)
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

	// Write
	err := db.SetGuildFeed(m.GuildID, entities.NewFeed(subreddit, title, author, pin, postType, m.ChannelID))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! This reddit feed has been saved. If there are valid posts they will start appearing in a few minutes."))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
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

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildFeeds := db.GetGuildFeeds(m.GuildID)

	// Check if there are set reddit feeds for this guild
	if guildFeeds == nil || len(guildFeeds) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error. There are no set reddit feeds.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	cmdStrs := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	if len(cmdStrs) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"removefeed [type]* [u/author]* [channel]* [r/subreddit] [title]*`\n\n* is optional\n\n"+
			"Type refers to the post sort filter. Valid values are `hot`, `new` and `rising`. Defaults to `hot`.\n"+
			"\nAuthor is the name of the post author.\n"+
			"\nChannel is the ID or name of a channel from which to remove\n"+
			"\nTitle is what a post title should start with or be for the BOT to post it. Leave empty for all feeds fulfilling [type] and [r/subreddit].")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
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
			common.LogError(s, guildSettings.BotLog, err)
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
		chaID, _ := common.ChannelParser(s, val, m.GuildID)
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

	// Fetches the target feed ID
	for _, feed := range guildFeeds {
		if feed.GetSubreddit() == subreddit  {
			if title != "" && feed.GetTitle() != title {
				continue
			}
			if author != "" && feed.GetAuthor() != author {
				continue
			}
			if postType != "" && feed.GetPostType() != postType {
				continue
			}
			if channelID != "" && feed.GetChannelID() != channelID {
				continue
			}
			break
		}
	}

	// Write
	err := db.SetGuildFeed(m.GuildID, entities.NewFeed(subreddit, title, author, false, postType, channelID), true)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! This reddit feed has been removed."))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Prints all currently set reddit feeds
func viewRedditFeedCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		message      string
		splitMessage []string
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildFeeds := db.GetGuildFeeds(m.GuildID)

	if guildFeeds == nil || len(guildFeeds) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set reddit feeds.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Iterates through all the reddit feeds if they exist and adds them to the message string and print them
	for _, feed := range guildFeeds {
		// Format print string
		message += fmt.Sprintf("**r/%s**", feed.GetSubreddit())
		if feed.GetAuthor() != "" {
			message += fmt.Sprintf(" - **u/%s**", feed.GetAuthor())
		}
		message += fmt.Sprintf(" - **%s**", feed.GetPostType())
		if feed.GetPin() {
			message += " - **pinned**"
		}
		message += fmt.Sprintf(" - **%s**", feed.GetChannelID())
		if feed.GetTitle() != "" {
			message += fmt.Sprintf(" - **%s**", feed.GetTitle())
		}
		message += "\n"
	}

	// Splits the message if it's over 1900 characters
	if len(message) > 1900 {
		splitMessage = common.SplitLongMessage(message)
	}

	// Prints split or unsplit whois
	if splitMessage == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	for i := 0; i < len(splitMessage); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot send feed message.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}
}

func init() {
	Add(&Command{
		Execute:    setRedditFeedCommand,
		Trigger:    "addfeed",
		Aliases:    []string{"setfeed", "adfeed", "addreddit", "setreddit"},
		Desc:       "Adds a reddit feed to the channel",
		Permission: functionality.Mod,
		Module:     "reddit",
	})
	Add(&Command{
		Execute:    removeRedditFeedCommand,
		Trigger:    "removefeed",
		Aliases:    []string{"killfeed", "deletefeed", "removereddit", "killreddit", "deletereddit"},
		Desc:       "Removes a reddit feed",
		Permission: functionality.Mod,
		Module:     "reddit",
	})
	Add(&Command{
		Execute:    viewRedditFeedCommand,
		Trigger:    "feeds",
		Aliases:    []string{"showreddit", "redditview", "redditshow", "printfeed", "viewfeeds", "showfeeds", "showfeed", "viewfeed", "feed"},
		Desc:       "Prints all currently set Reddit feeds",
		Permission: functionality.Mod,
		Module:     "reddit",
	})
}
