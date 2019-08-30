package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Sets an RSS by subreddit and other params
func setRssCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		subreddit string
		author    string
		postType  = "hot"
		pin       bool
		title     string

		subIndex int
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	cmdStrs := strings.Split(messageLowercase, " ")

	if len(cmdStrs) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"setrss [u/author]* [type]* [pin]* [r/subreddit] [title]*`\n\n * are optional.\n\nType refers to the post sort filter. Valid values are `hot`, `new` and `rising`. Defaults to `hot`.\nPin refers to whether to pin the post when the bot posts it and unpin the previous bot pin of the same subreddit. Use `true` or `false` as values.\nTitle is what a post title should start with for the BOT to post it. Leave empty for all posts.\n\nFor author and subreddit be sure to add the prefixes `u/` and `r/`.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
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
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
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

	misc.MapMutex.Lock()
	err := misc.RssThreadsWrite(subreddit, author, title, postType, m.ChannelID, m.GuildID, pin)
	if err != nil {
		misc.MapMutex.Unlock()
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	misc.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! This RSS setting has been saved."))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Removes a previously set RSS
func removeRssCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		subreddit 	string
		title 		string
		postType 	string
		author		string
		channelID	string

		subIndex	int
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	// Check if there are set RSS settings
	if len(misc.GuildMap[m.GuildID].RssThreads) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error. There are no set rss threads.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	cmdStrs := strings.Split(messageLowercase, " ")

	if len(cmdStrs) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"removerss [type]* [u/author]* [channel]* [r/subreddit] [title]*`\n\n* is optional\n\n" +
			"Type refers to the post sort filter. Valid values are `hot`, `new` and `rising`. Defaults to `hot`.\n" +
			"\nAuthor is the name of the post author.\n" +
			"\nChannel is the ID or name of a channel from which to remove\n" +
			"\nTitle is what a post title should start with or be for the BOT to post it. Leave empty for all RSS settings fulfilling [type] and [r/subreddit].")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
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
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
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
		chaID, _ := misc.ChannelParser(s, val, m.GuildID)
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

	misc.MapMutex.Lock()
	err := misc.RssThreadsRemove(subreddit, title, author, postType, channelID, m.GuildID)
	if err != nil {
		misc.MapMutex.Unlock()
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	misc.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! This RSS setting has been removed."))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Prints all currently set RSS
func viewRssCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		message		 string
		splitMessage []string
	)

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	if len(misc.GuildMap[m.GuildID].RssThreads) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set RSS threads.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	// Iterates through all the rss settings if they exist and adds them to the message string and print them
	for i := 0; i < len(misc.GuildMap[m.GuildID].RssThreads); i++ {
		// Format print string
		message += fmt.Sprintf("`r/%v", misc.GuildMap[m.GuildID].RssThreads[i].Subreddit)
		if misc.GuildMap[m.GuildID].RssThreads[i].Author != "" {
			message += fmt.Sprintf(" - u/%v", misc.GuildMap[m.GuildID].RssThreads[i].Author)
		}
		message += fmt.Sprintf(" - %v", misc.GuildMap[m.GuildID].RssThreads[i].PostType)
		if misc.GuildMap[m.GuildID].RssThreads[i].Pin {
			message += " - pinned"
		}
		message += fmt.Sprintf(" - %v", misc.GuildMap[m.GuildID].RssThreads[i].ChannelID)
		if misc.GuildMap[m.GuildID].RssThreads[i].Title != "" {
			message += fmt.Sprintf(" - %v", misc.GuildMap[m.GuildID].RssThreads[i].Title)
		}
		message += "`\n"
	}
	misc.MapMutex.Unlock()

	// Splits the message if it's over 1900 characters
	if len(message) > 1900 {
		splitMessage = misc.SplitLongMessage(message)
	}

	// Prints split or unsplit whois
	if splitMessage == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	for i := 0; i < len(splitMessage); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot send rss message.")
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
		}
	}
}

func init() {
	add(&command{
		execute:  setRssCommand,
		trigger:  "setrss",
		aliases: []string{"addrss"},
		desc:     "Assigns a reddit RSS to the channel.",
		elevated: true,
		category: "rss",
	})
	add(&command{
		execute:  removeRssCommand,
		trigger:  "removerss",
		aliases: []string{"killrss", "deleterss"},
		desc:     "Removes a previously set reddit RSS.",
		elevated: true,
		category: "rss",
	})
	add(&command{
		execute:  viewRssCommand,
		trigger:  "viewrss",
		aliases:  []string{"showrss", "rssview", "rssshow", "viewrs", "showrs", "rss"},
		desc:     "Prints all currently set reddit RSS.",
		elevated: true,
		category: "rss",
	})
}
