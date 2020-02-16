package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"ZeroTsu/functionality"
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

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"killchannel [channel]`")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fetches channel ID
	channelID, channelName = functionality.ChannelParser(s, commandStrings[1], m.GuildID)
	if channelID == "" && channelName == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Removes the reddit feeds for that channel
	functionality.Mutex.Lock()
	for rssLoopFlag {
		if rssTimerFlag {
			for _, rssTimer := range functionality.GuildMap[m.GuildID].FeedChecks {
				if rssTimer.Thread.ChannelID == channelID {
					rssTimerFlag = true
					err := functionality.RssThreadsTimerRemove(rssTimer.Thread, m.GuildID)
					if err != nil {
						functionality.Mutex.Unlock()
						functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
						return
					}
					break
				} else {
					rssTimerFlag = false
				}
			}
			if len(functionality.GuildMap[m.GuildID].FeedChecks) == 0 {
				rssTimerFlag = false
			}
		}

		for _, thread := range functionality.GuildMap[m.GuildID].Feeds {
			if thread.ChannelID == channelID {
				rssLoopFlag = true
				err := functionality.RssThreadsRemove(thread.Subreddit, thread.Title, thread.Author, thread.PostType, thread.ChannelID, m.GuildID)
				if err != nil {
					functionality.Mutex.Unlock()
					functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
					return
				}
				break
			} else {
				rssLoopFlag = false
			}
		}
		if len(functionality.GuildMap[m.GuildID].Feeds) == 0 {
			rssLoopFlag = false
		}
	}
	functionality.Mutex.Unlock()

	// Fixes role name bug by hyphenating the channel name
	roleName = strings.Replace(strings.TrimSpace(channelName), " ", "-", -1)
	roleName = strings.Replace(roleName, "--", "-", -1)

	// Fetches channel role ID by finding it amongst all server roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	for _, role := range roles {
		if strings.ToLower(role.Name) == strings.ToLower(roleName) {
			roleID = role.ID
			break
		}
	}

	// Deletes all set reacts that link to the role ID if not using Kaguya
	functionality.Mutex.RLock()
	for messageID, roleMapMap := range functionality.GuildMap[m.GuildID].ReactJoinMap {
		for _, roleEmojiMap := range roleMapMap.RoleEmojiMap {
			for role, emojiSlice := range roleEmojiMap {
				if strings.ToLower(role) == strings.ToLower(roleName) {
					for _, emoji := range emojiSlice {
						// Remove React Join command
						author.ID = s.State.User.ID
						message.ID = messageID
						message.GuildID = m.GuildID
						message.Author = &author
						message.Content = fmt.Sprintf("%sremovereact %s %s", guildSettings.Prefix, messageID, emoji)
						functionality.Mutex.RUnlock()
						removeReactJoinCommand(s, &message)
						functionality.Mutex.RLock()
					}
				}
			}
		}
	}
	functionality.Mutex.RUnlock()

	// Removes the role
	if roleID != "" {
		err = s.GuildRoleDelete(m.GuildID, roleID)
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	// Removes the channel
	_, err = s.ChannelDelete(channelID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err = s.ChannelMessageSend(m.ChannelID, "Success: Channel `"+channelName+"` was successfuly deleted!")
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.Unlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"killcategory [category]`")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fetches category ID
	categoryID, categoryName = functionality.CategoryParser(s, commandStrings[1], m.GuildID)
	if categoryID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such category exists.")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	if strings.ToLower(categoryName) == "general" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Not allowed to delete the general category. Please try something else.")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Starting channel deletion. For categories with a lot of channels you will have to wait more. A message will be sent when it is done.")
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	for loopFlag {
		channels, err := s.GuildChannels(m.GuildID)
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
				message.Content = fmt.Sprintf("%vkillchannel %v", guildSettings.Prefix, channel.ID)
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
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err = s.ChannelMessageSend(m.ChannelID, "Success: Module `"+categoryName+"` was successfuly deleted!")
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"killchannelreacts [channel]`")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fetches channel ID
	channelID, channelName = functionality.ChannelParser(s, commandStrings[1], m.GuildID)
	if channelID == "" && channelName == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fixes role name bug by hyphenating the channel name
	roleName = strings.Replace(strings.TrimSpace(channelName), " ", "-", -1)
	roleName = strings.Replace(roleName, "--", "-", -1)

	// Deletes all set reacts that link to the role ID if not using Kaguya
	functionality.Mutex.RLock()
	for messageID, roleMapMap := range functionality.GuildMap[m.GuildID].ReactJoinMap {
		for _, roleEmojiMap := range roleMapMap.RoleEmojiMap {
			for role, emojiSlice := range roleEmojiMap {
				if strings.ToLower(role) == strings.ToLower(roleName) {
					for _, emoji := range emojiSlice {
						// Remove React Join command
						author.ID = s.State.User.ID
						message.ID = messageID
						message.GuildID = m.GuildID
						message.Author = &author
						message.Content = fmt.Sprintf("%sremovereact %s %s", guildSettings.Prefix, messageID, emoji)
						functionality.Mutex.RUnlock()
						removeReactJoinCommand(s, &message)
						functionality.Mutex.RLock()
					}
				}
			}
		}
	}
	functionality.Mutex.RUnlock()

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err := s.ChannelMessageSend(m.ChannelID, "Success: Channel `"+channelName+"`'s set react joins were removed!")
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"killcategoryreacts [category]`")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fetches category ID
	categoryID, categoryName = functionality.CategoryParser(s, commandStrings[1], m.GuildID)
	if categoryID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such category exists.")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Starting channel react deletion. For categories with a lot of channels you will have to wait more. A message will be sent when it is done.")
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	channels, err := s.GuildChannels(m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	for _, channel := range channels {
		if channel.ParentID == categoryID {
			// Delete channel reacts Command
			author.ID = s.State.User.ID
			message.GuildID = m.GuildID
			message.Author = &author
			message.ChannelID = m.ChannelID
			message.Content = fmt.Sprintf("%vkillchannelreacts %v", guildSettings.Prefix, channel.ID)
			deleteChannelReacts(s, &message)
		}
	}

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err = s.ChannelMessageSend(m.ChannelID, "Success: Module `"+categoryName+"`'s set react joins were removed!")
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    deleteChannel,
		Trigger:    "killchannel",
		Aliases:    []string{"deletechannel", "removechannel"},
		Desc:       "Removes a channel, its role, and all associated reacts and Reddit feeds",
		Permission: functionality.Mod,
		Module:     "channel",
	})
	functionality.Add(&functionality.Command{
		Execute:    deleteCategory,
		Trigger:    "killcategory",
		Aliases:    []string{"deletecategory", "removecategory"},
		Desc:       "Removes every channel in a category, their roles, and all associated reacts and Reddit feeds",
		Permission: functionality.Mod,
		Module:     "misc",
	})
	functionality.Add(&functionality.Command{
		Execute:    deleteChannelReacts,
		Trigger:    "killchannelreacts",
		Aliases:    []string{"removechannelreacts", "removechannelreact", "killchannelreact", "deletechannelreact", "deletechannelreacts"},
		Desc:       "Removes all reacts linked to a specific channel [REACTS]",
		Permission: functionality.Mod,
		Module:     "reacts",
	})
	functionality.Add(&functionality.Command{
		Execute:    deleteCategoryReacts,
		Trigger:    "killcategoryreacts",
		Aliases:    []string{"removecategoryreacts", "removecategoryreact", "killcategoryreact", "deletecategoryreact", "deletecategoryreacts"},
		Desc:       "Removes all reacts linked to a specific category [REACTS]",
		Permission: functionality.Mod,
		Module:     "reacts",
	})
}
