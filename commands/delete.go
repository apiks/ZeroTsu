package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Deletes a channel's set rss, reacts linked to them and their role
func deleteChannel(s *discordgo.Session, m *discordgo.Message) {
	var (
		channelID    string
		channelName  string
		roleName     string
		roleID       string
		rssLoopFlag  = true
		rssTimerFlag = true

		message discordgo.Message
		author  discordgo.User
	)

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"killchannel [channel]`")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fetches channel ID
	channelID, channelName = common.ChannelParser(s, commandStrings[1], m.GuildID)
	if channelID == "" && channelName == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Removes the reddit feeds for that channel
	guildFeeds := db.GetGuildFeeds(m.GuildID)
	guildFeedChecks := db.GetGuildFeedChecks(m.GuildID)
	for rssLoopFlag {
		if rssTimerFlag {
			for _, feedCheck := range guildFeedChecks {
				if feedCheck == nil {
					continue
				}

				if feedCheck.GetFeed().GetChannelID() == channelID {
					rssTimerFlag = true
					err := db.SetGuildFeedCheck(m.GuildID, feedCheck, true)
					if err != nil {
						common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
						return
					}
					break
				} else {
					rssTimerFlag = false
				}
			}
			if len(guildFeedChecks) == 0 {
				rssTimerFlag = false
			}
		}

		for _, feed := range guildFeeds {
			if feed == nil {
				return
			}

			if feed.GetChannelID() == channelID {
				rssLoopFlag = true
				err := db.SetGuildFeed(m.GuildID, feed, true)
				if err != nil {
					common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
					return
				}
				break
			} else {
				rssLoopFlag = false
			}
		}
		if len(guildFeeds) == 0 {
			rssLoopFlag = false
		}
	}

	// Fixes role name bug by hyphenating the channel name
	roleName = strings.Replace(strings.TrimSpace(channelName), " ", "-", -1)
	roleName = strings.Replace(roleName, "--", "-", -1)

	// Fetches channel role ID by finding it amongst all server roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	for _, role := range roles {
		if strings.ToLower(role.Name) == strings.ToLower(roleName) {
			roleID = role.ID
			break
		}
	}

	// Deletes all set reacts that link to the role ID if not using Kaguya
	reactJoins := db.GetGuildReactJoin(m.GuildID)
	for messageID, roleMapMap := range reactJoins {
		if roleMapMap == nil {
			continue
		}

		for _, roleEmojiMap := range roleMapMap.GetRoleEmojiMap() {
			if roleEmojiMap == nil {
				continue
			}

			for role, emojiSlice := range roleEmojiMap {
				if emojiSlice == nil {
					continue
				}

				if strings.ToLower(role) == strings.ToLower(roleName) {
					for _, emoji := range emojiSlice {
						// Remove React Join command
						author.ID = s.State.User.ID
						message.ID = messageID
						message.GuildID = m.GuildID
						message.Author = &author
						message.Content = fmt.Sprintf("%sremovereact %s %s", guildSettings.GetPrefix(), messageID, emoji)
						removeReactJoinCommand(s, &message)
					}
				}
			}
		}
	}

	// Removes the role
	if roleID != "" {
		err = s.GuildRoleDelete(m.GuildID, roleID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	// Removes the channel
	_, err = s.ChannelDelete(channelID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err = s.ChannelMessageSend(m.ChannelID, "Success: Channel `"+channelName+"` was successfuly deleted!")
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// Deletes all of the channels of a category, their set rss, reacts linked to them and their roles and deletes the category
func deleteCategory(s *discordgo.Session, m *discordgo.Message) {
	var (
		categoryID   string
		categoryName string
		loopFlag     = true

		message discordgo.Message
		author  discordgo.User
	)

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"killcategory [category]`")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fetches category ID
	categoryID, categoryName = common.CategoryParser(s, commandStrings[1], m.GuildID)
	if categoryID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such category exists.")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	if strings.ToLower(categoryName) == "general" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Not allowed to delete the general category. Please try something else.")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Starting channel deletion. For categories with a lot of channels you will have to wait more. A message will be sent when it is done.")
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	for loopFlag {
		channels, err := s.GuildChannels(m.GuildID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		for _, channel := range channels {
			if channel.ParentID == categoryID {
				loopFlag = true
				// Delete channel Command
				author.ID = s.State.User.ID
				message.GuildID = m.GuildID
				message.Author = &author
				message.ChannelID = m.ChannelID
				message.Content = fmt.Sprintf("%vkillchannel %v", guildSettings.GetPrefix(), channel.ID)
				deleteChannel(s, &message)
				break
			} else {
				loopFlag = false
			}
		}
		if len(channels) == 0 {
			loopFlag = false
		}
	}

	_, err = s.ChannelDelete(categoryID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err = s.ChannelMessageSend(m.ChannelID, "Success: Module `"+categoryName+"` was successfuly deleted!")
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// Deletes all reacts linked to a specific channel
func deleteChannelReacts(s *discordgo.Session, m *discordgo.Message) {
	var (
		channelID   string
		channelName string
		roleName    string

		message discordgo.Message
		author  discordgo.User
	)

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"killchannelreacts [channel]`")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fetches channel ID
	channelID, channelName = common.ChannelParser(s, commandStrings[1], m.GuildID)
	if channelID == "" && channelName == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fixes role name bug by hyphenating the channel name
	roleName = strings.Replace(strings.TrimSpace(channelName), " ", "-", -1)
	roleName = strings.Replace(roleName, "--", "-", -1)

	// Deletes all set reacts that link to the role ID if not using Kaguya
	reactJoins := db.GetGuildReactJoin(m.GuildID)
	for messageID, roleMapMap := range reactJoins {
		if roleMapMap == nil {
			continue
		}

		for _, roleEmojiMap := range roleMapMap.GetRoleEmojiMap() {
			if roleEmojiMap == nil {
				continue
			}

			for role, emojiSlice := range roleEmojiMap {
				if emojiSlice == nil {
					continue
				}

				if strings.ToLower(role) == strings.ToLower(roleName) {
					for _, emoji := range emojiSlice {
						// Remove React Join command
						author.ID = s.State.User.ID
						message.ID = messageID
						message.GuildID = m.GuildID
						message.Author = &author
						message.Content = fmt.Sprintf("%sremovereact %s %s", guildSettings.GetPrefix(), messageID, emoji)
						removeReactJoinCommand(s, &message)
					}
				}
			}
		}
	}

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err := s.ChannelMessageSend(m.ChannelID, "Success: Channel `"+channelName+"`'s set react joins were removed!")
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// Deletes all reacts linked to the channels of a specific category
func deleteCategoryReacts(s *discordgo.Session, m *discordgo.Message) {
	var (
		categoryID   string
		categoryName string

		message discordgo.Message
		author  discordgo.User
	)

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"killcategoryreacts [category]`")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fetches category ID
	categoryID, categoryName = common.CategoryParser(s, commandStrings[1], m.GuildID)
	if categoryID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such category exists.")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Starting channel react deletion. For categories with a lot of channels you will have to wait more. A message will be sent when it is done.")
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	channels, err := s.GuildChannels(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	for _, channel := range channels {
		if channel.ParentID == categoryID {
			// Delete channel reacts Command
			author.ID = s.State.User.ID
			message.GuildID = m.GuildID
			message.Author = &author
			message.ChannelID = m.ChannelID
			message.Content = fmt.Sprintf("%vkillchannelreacts %v", guildSettings.GetPrefix(), channel.ID)
			deleteChannelReacts(s, &message)
		}
	}

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err = s.ChannelMessageSend(m.ChannelID, "Success: Module `"+categoryName+"`'s set react joins were removed!")
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

func init() {
	Add(&Command{
		Execute:    deleteChannel,
		Trigger:    "killchannel",
		Aliases:    []string{"deletechannel", "removechannel"},
		Desc:       "Removes a channel, its role, and all associated reacts and Reddit feeds",
		Permission: functionality.Mod,
		Module:     "channel",
	})
	Add(&Command{
		Execute:    deleteCategory,
		Trigger:    "killcategory",
		Aliases:    []string{"deletecategory", "removecategory"},
		Desc:       "Removes every channel in a category, their roles, and all associated reacts and Reddit feeds",
		Permission: functionality.Mod,
		Module:     "misc",
	})
	Add(&Command{
		Execute:    deleteChannelReacts,
		Trigger:    "killchannelreacts",
		Aliases:    []string{"removechannelreacts", "removechannelreact", "killchannelreact", "deletechannelreact", "deletechannelreacts"},
		Desc:       "Removes all reacts linked to a specific channel [REACTS]",
		Permission: functionality.Mod,
		Module:     "reacts",
	})
	Add(&Command{
		Execute:    deleteCategoryReacts,
		Trigger:    "killcategoryreacts",
		Aliases:    []string{"removecategoryreacts", "removecategoryreact", "killcategoryreact", "deletecategoryreact", "deletecategoryreacts"},
		Desc:       "Removes all reacts linked to a specific category [REACTS]",
		Permission: functionality.Mod,
		Module:     "reacts",
	})
}
