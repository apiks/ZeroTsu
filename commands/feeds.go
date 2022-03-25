package commands

import (
	"fmt"
	"strings"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// addRedditFeedCommand adds a reddit feed by subreddit and other args to a channel
func addRedditFeedCommand(targetChannel *discordgo.Channel, subreddit, author, postType, title string, pin bool) string {
	if strings.HasPrefix(subreddit, "r/") || strings.HasPrefix(subreddit, "/r/") {
		subreddit = strings.TrimPrefix(subreddit, "/r/")
		subreddit = strings.TrimPrefix(subreddit, "r/")
	}
	if strings.HasPrefix(author, "u/") || strings.HasPrefix(author, "/u/") {
		author = strings.TrimPrefix(author, "/u/")
		author = strings.TrimPrefix(author, "u/")
	}
	if postType != "hot" && postType != "rising" && postType != "new" {
		return "Error: Invalid post type."
	}

	err := db.SetGuildFeed(targetChannel.GuildID, entities.NewFeed(subreddit, title, author, pin, postType, targetChannel.ID))
	if err != nil {
		return err.Error()
	}

	return "Success! This reddit feed has been added. If there are valid posts they will start appearing within an hour or two."
}

// addRedditFeedCommandHandler adds a reddit feed by subreddit and other args to a channel
func addRedditFeedCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! This reddit feed has been saved. If there are valid posts they will start appearing within an hour or two."))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// removeRedditFeedCommand removes a previously set reddit feed
func removeRedditFeedCommand(targetChannel *discordgo.Channel, subreddit, author, postType, title string) string {
	guildFeeds := db.GetGuildFeeds(targetChannel.GuildID)

	if guildFeeds == nil || len(guildFeeds) == 0 {
		return "Error. There are no set reddit feeds."
	}

	if strings.HasPrefix(subreddit, "/r/") || strings.HasPrefix(subreddit, "r/") {
		subreddit = strings.TrimPrefix(subreddit, "/r/")
		subreddit = strings.TrimPrefix(subreddit, "r/")
	}
	if strings.HasPrefix(author, "/u/") || strings.HasPrefix(author, "u/") {
		author = strings.TrimPrefix(author, "/u/")
		author = strings.TrimPrefix(author, "u/")
	}
	if postType != "hot" && postType != "rising" && postType != "new" {
		return "Error: Invalid post type."
	}

	// Write
	err := db.SetGuildFeed(targetChannel.GuildID, entities.NewFeed(subreddit, title, author, false, postType, targetChannel.ID), true)
	if err != nil {
		return err.Error()
	}

	return "Success! This reddit feed has been removed."
}

// removeRedditFeedCommandHandler removes a previously set reddit feed
func removeRedditFeedCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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
		if feed.GetSubreddit() == subreddit {
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

// viewRedditFeedCommand prints all currently set reddit feeds
func viewRedditFeedCommand(guildID string) []string {
	var (
		message    string
		guildFeeds = db.GetGuildFeeds(guildID)
	)

	if guildFeeds == nil || len(guildFeeds) == 0 {
		return []string{"Error: There are no set reddit feeds."}
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

	return common.SplitLongMessage(message)
}

// viewRedditFeedCommandHandler prints all currently set reddit feeds
func viewRedditFeedCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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
		Execute:    addRedditFeedCommandHandler,
		Name:       "add-reddit-feed",
		Aliases:    []string{"setfeed", "adfeed", "addreddit", "setreddit", "addredditfeed"},
		Desc:       "Adds a reddit feed to a channel",
		Permission: functionality.Mod,
		Module:     "reddit",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel in which you want add a reddit feed in.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "subreddit",
				Description: "The subreddit you want to set a feed for. Quarantined or private subreddits do not work.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "author",
				Description: "Filter to a user who you want posts only from.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "post-type",
				Description: "The type of feed filter you want to set. Defaults to 'hot'. Use 'hot', 'rising' or 'new'.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "title",
				Description: "Filter to posts starting only with this title.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "pin",
				Description: "Whether to automatically pin the latest post and unpin the previous one.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "add-reddit-feed", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			var (
				targetChannel *discordgo.Channel
				subreddit     string
				author        string
				postType      = "hot"
				title         string
				pin           bool
			)
			if i.ApplicationCommandData().Options == nil {
				return
			}
			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "channel" {
					targetChannel = option.ChannelValue(s)
				} else if option.Name == "subreddit" {
					subreddit = option.StringValue()
				} else if option.Name == "author" {
					author = option.StringValue()
				} else if option.Name == "post-type" {
					postType = option.StringValue()
				} else if option.Name == "title" {
					title = option.StringValue()
				} else if option.Name == "pin" {
					pin = option.BoolValue()
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: addRedditFeedCommand(targetChannel, subreddit, author, postType, title, pin),
				},
			})
		},
	})
	Add(&Command{
		Execute:    removeRedditFeedCommandHandler,
		Name:       "remove-reddit-feed",
		Aliases:    []string{"killfeed", "deletefeed", "removereddit", "killreddit", "deletereddit", "removeredditfeed"},
		Desc:       "Removes a reddit feed from a channel",
		Permission: functionality.Mod,
		Module:     "reddit",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel in from which you want to remove reddit feeds.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "subreddit",
				Description: "The subreddit you want to remove.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "author",
				Description: "The author filter the feed you want to remove has.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "post-type",
				Description: "The post type filter the feed you want to remove has.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "title",
				Description: "The title filter the feed you want to remove has.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "remove-reddit-feed", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			var (
				targetChannel *discordgo.Channel
				subreddit     string
				author        string
				postType      = "hot"
				title         string
			)
			if i.ApplicationCommandData().Options == nil {
				return
			}
			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "channel" {
					targetChannel = option.ChannelValue(s)
				} else if option.Name == "subreddit" {
					subreddit = option.StringValue()
				} else if option.Name == "author" {
					author = option.StringValue()
				} else if option.Name == "post-type" {
					postType = strings.ToLower(option.StringValue())
				} else if option.Name == "title" {
					title = option.StringValue()
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: removeRedditFeedCommand(targetChannel, subreddit, author, postType, title),
				},
			})
		},
	})
	Add(&Command{
		Execute:    viewRedditFeedCommandHandler,
		Name:       "reddit-feeds",
		Aliases:    []string{"showreddit", "redditview", "redditshow", "printfeed", "viewfeeds", "showfeeds", "showfeed", "viewfeed", "feed", "feeds", "redditfeeds"},
		Desc:       "Prints all currently set Reddit feeds",
		Permission: functionality.Mod,
		Module:     "reddit",
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "reddit-feeds", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			messages := viewRedditFeedCommand(i.GuildID)
			if messages == nil {
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: messages[0],
				},
			})

			if len(messages) > 1 {
				for j, message := range messages {
					if j == 0 {
						continue
					}

					s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
						Content: message,
					})
				}
			}
		},
	})
}
