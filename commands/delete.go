package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Deletes a channel's set rss, reacts linked to them and their role
func deleteChannel(s *discordgo.Session, m *discordgo.Message) {
	var (
		channelID 		string
		channelName 	string
		roleName		string
		roleID			string
		rssLoopFlag		= true
		rssTimerFlag	= true

		message 		discordgo.Message
		author  		discordgo.User
	)

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "killchannel [channel]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fetches channel ID
	channelID, channelName = misc.ChannelParser(s, commandStrings[1])
	if channelID == "" && channelName == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Removes the set RSS feeds for that channel
	misc.MapMutex.Lock()
	for rssLoopFlag {
		if rssTimerFlag {
			for _, rssTimer := range misc.ReadRssThreadsCheck {
				if rssTimer.ChannelID == channelID {
					rssTimerFlag = true
					err := misc.RssThreadsTimerRemove(rssTimer.Thread, rssTimer.Date, rssTimer.ChannelID)
					if err != nil {
						misc.MapMutex.Unlock()
						misc.CommandErrorHandler(s, m, err)
						return
					}
					break
				} else {
					rssTimerFlag = false
				}
			}
			if len (misc.ReadRssThreadsCheck) == 0 {
				rssTimerFlag = false
			}
		}

		for _, thread := range misc.ReadRssThreads {
			if thread.Channel == channelID {
				rssLoopFlag = true
				_, err := misc.RssThreadsRemove(thread.Thread, thread.Author)
				if err != nil {
					misc.MapMutex.Unlock()
					misc.CommandErrorHandler(s, m, err)
					return
				}
				break
			} else {
				rssLoopFlag = false
			}
		}
		if len(misc.ReadRssThreads) == 0 {
			rssLoopFlag = false
		}
	}
	misc.MapMutex.Unlock()

	// Fixes role name bug by hyphenating the channel name
	roleName = strings.Replace(strings.TrimSpace(channelName), " ", "-", -1)
	roleName = strings.Replace(roleName, "--", "-", -1)

	// Fetches channel role ID by finding it amongst all server roles
	roles, err := s.GuildRoles(config.ServerID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
	for _, role := range roles {
		if strings.ToLower(role.Name) == strings.ToLower(roleName) {
			roleID = role.ID
			break
		}
	}

	// Deletes all set reacts that link to the role ID if not using Kaguya
	misc.MapMutex.Lock()
	for messageID, roleMapMap := range reactChannelJoinMap {
		for _, roleEmojiMap := range roleMapMap.RoleEmojiMap {
			for role, emojiSlice := range roleEmojiMap {
				if strings.ToLower(role) == strings.ToLower(roleName) {
					for _, emoji := range emojiSlice {
						// Remove React Join command
						author.ID = s.State.User.ID
						message.ID = messageID
						message.Author = &author
						message.Content = fmt.Sprintf("%vremovereact %v %v", config.BotPrefix, messageID, emoji)
						misc.MapMutex.Unlock()
						removeReactJoinCommand(s, &message)
						misc.MapMutex.Lock()
					}
				}
			}
		}
	}
	misc.MapMutex.Unlock()

	// Removes the role
	if roleID != "" {
		err = s.GuildRoleDelete(config.ServerID, roleID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}
	}

	// Removes the channel
	_, err = s.ChannelDelete(channelID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err = s.ChannelMessageSend(m.ChannelID, "Success: Channel `" + channelName + "` was successfuly deleted!")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Deletes all of the channels of a category, their set rss, reacts linked to them and their roles and deletes the category
func deleteCategory(s *discordgo.Session, m *discordgo.Message) {
	var (
		categoryID 		string
		categoryName	string
		loopFlag		= true

		message 		discordgo.Message
		author  		discordgo.User
	)

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "killcategory [category]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fetches category ID
	categoryID, categoryName = misc.CategoryParser(s, commandStrings[1])
	if categoryID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such category exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	if strings.ToLower(categoryName) == "general" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Not allowed to delete the general category. Please try something else.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Starting channel deletion. For categories with a lot of channels you will have to wait more. A message will be sent when it is done.")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	for loopFlag {
		channels, err := s.GuildChannels(config.ServerID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}
		for _, channel := range channels {
			if channel.ParentID == categoryID {
				loopFlag = true
				// Delete channel Command
				author.ID = s.State.User.ID
				message.Author = &author
				message.ChannelID = m.ChannelID
				message.Content = fmt.Sprintf("%vkillchannel %v", config.BotPrefix, channel.ID)
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
		misc.CommandErrorHandler(s, m, err)
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err = s.ChannelMessageSend(m.ChannelID, "Success: Category `" + categoryName + "` was successfuly deleted!")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Deletes all reacts linked to a specific channel
func deleteChannelReacts(s *discordgo.Session, m *discordgo.Message) {
	var (
		channelID 		string
		channelName 	string
		roleName		string

		message 		discordgo.Message
		author  		discordgo.User
	)

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "killchannelreacts [channel]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fetches channel ID
	channelID, channelName = misc.ChannelParser(s, commandStrings[1])
	if channelID == "" && channelName == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such channel exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fixes role name bug by hyphenating the channel name
	roleName = strings.Replace(strings.TrimSpace(channelName), " ", "-", -1)
	roleName = strings.Replace(roleName, "--", "-", -1)

	// Deletes all set reacts that link to the role ID if not using Kaguya
	misc.MapMutex.Lock()
	for messageID, roleMapMap := range reactChannelJoinMap {
		for _, roleEmojiMap := range roleMapMap.RoleEmojiMap {
			for role, emojiSlice := range roleEmojiMap {
				if strings.ToLower(role) == strings.ToLower(roleName) {
					for _, emoji := range emojiSlice {
						// Remove React Join command
						author.ID = s.State.User.ID
						message.ID = messageID
						message.Author = &author
						message.Content = fmt.Sprintf("%vremovereact %v %v", config.BotPrefix, messageID, emoji)
						misc.MapMutex.Unlock()
						removeReactJoinCommand(s, &message)
						misc.MapMutex.Lock()
					}
				}
			}
		}
	}
	misc.MapMutex.Unlock()

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err := s.ChannelMessageSend(m.ChannelID, "Success: Channel `" + channelName + "`'s set react joins were removed!")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Deletes all reacts linked to the channels of a specific category
func deleteCategoryReacts(s *discordgo.Session, m *discordgo.Message) {
	var (
		categoryID 		string
		categoryName	string

		message 		discordgo.Message
		author  		discordgo.User
	)

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "killcategoryreacts [category]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fetches category ID
	categoryID, categoryName = misc.ChannelParser(s, commandStrings[1])
	if categoryID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such category exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Starting channel react deletion. For categories with a lot of channels you will have to wait more. A message will be sent when it is done.")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	channels, err := s.GuildChannels(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}
	for _, channel := range channels {
		if channel.ParentID == categoryID {
			// Delete channel reacts Command
			author.ID = s.State.User.ID
			message.Author = &author
			message.ChannelID = m.ChannelID
			message.Content = fmt.Sprintf("%vkillchannelreacts %v", config.BotPrefix, channel.ID)
			deleteChannelReacts(s, &message)
		}
	}

	if m.Author.ID == s.State.User.ID {
		return
	}
	_, err = s.ChannelMessageSend(m.ChannelID, "Success: Category `" + categoryName + "`'s set react joins were removed!")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func init() {
	add(&command{
		execute: deleteChannel,
		trigger: "killchannel",
		aliases: []string{"deletechannel", "removechannel"},
		desc:    "Removes a channel, its role, and all associated reacts and RSS feeds.",
		elevated: true,
		category:"channel",
	})
	add(&command{
		execute: deleteCategory,
		trigger: "killcategory",
		aliases: []string{"deletecategory", "removecategory"},
		desc:    "Removes every channel in a category, their roles, and all associated reacts and RSS feeds.",
		elevated: true,
		category:"misc",
	})
	add(&command{
		execute:  deleteChannelReacts,
		trigger:  "killchannelreacts",
		aliases:  []string{"removechannelreacts", "removechannelreact", "killchannelreact", "deletechannelreact", "deletechannelreacts"},
		desc:     "Removes all reacts linked to a specific channel.",
		elevated: true,
		category: "reacts",
	})
	add(&command{
		execute:  deleteCategoryReacts,
		trigger:  "killcategoryreacts",
		aliases:  []string{"removecategoryreacts", "removecategoryreact", "killcategoryreact", "deletecategoryreact", "deletecategoryreacts"},
		desc:     "Removes all reacts linked to a specific category.",
		elevated: true,
		category: "reacts",
	})
}