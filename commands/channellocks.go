package commands

import (
	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Locks a specific channel and is spoiler-channel sensitive
func lockCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		roleExists       = false
		roleID           string
		roleTempPosition int
		spoilerRole 	 = false
	)

	// Pulls info on the channel the message is in
	cha, err := s.Channel(m.ChannelID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// If channel is already locked then notify user
	for index, permission := range cha.PermissionOverwrites {
		if permission.ID == config.ServerID {
			if cha.PermissionOverwrites[index].Deny == discordgo.PermissionSendMessages {
				_, err = s.ChannelMessageSend(m.ChannelID, "🔒 Channel already locked.")
				if err != nil {
					_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
					if err != nil {
						return
					}
					return
				}
				return
			}
		}
	}


	// Error if lock used in moderator category
	if cha.ParentID == "360994750796529665" {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Cannot lock a mod channel due to permission reasons.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fetches info on server roles from the server and puts it in deb
	deb, err := s.GuildRoles(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Updates opt-in-under and opt-in-above position
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == config.OptInUnder {
			misc.OptinUnderPosition = deb[i].Position
		} else if deb[i].Name == config.OptInAbove {
			misc.OptinAbovePosition = deb[i].Position
		} else if deb[i].Name == cha.Name {
			roleID = deb[i].ID
			roleTempPosition = deb[i].Position
		}
	}

	// Checks if the channel being locked is between the opt-ins
	for i := 0; i < len(deb); i++ {
		if roleTempPosition < misc.OptinUnderPosition &&
			roleTempPosition > misc.OptinAbovePosition {
			spoilerRole = true
			break
		}
	}

	if spoilerRole {
		// Removes send permissions only from the channel role if it's a spoiler channel
		err = s.ChannelPermissionSet(m.ChannelID, roleID, "role", discordgo.PermissionReadMessages, discordgo.PermissionSendMessages)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}
	} else {
		// Removes send permission from @everyone
		err = s.ChannelPermissionSet(m.ChannelID, config.ServerID, "role", 0, discordgo.PermissionSendMessages)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}
	}

	// Checks if the channel has a permission overwrite for all mod positions
	for i := 0; i < len(cha.PermissionOverwrites); i++ {
		for _, goodRole := range config.CommandRoles {
			if cha.PermissionOverwrites[i].ID == goodRole {
				roleExists = true
				break
			}
		}
		if roleExists {
			break
		}
	}

	// If the mod permission overwrite doesn't exist it adds it
	if !roleExists {
		for i := 0; i < len(deb); i++ {
			for _, goodRole := range config.CommandRoles {
				err = s.ChannelPermissionSet(m.ChannelID, goodRole, "role", discordgo.PermissionAll, 0)
				if err != nil {
					misc.CommandErrorHandler(s, m, err)
					return
				}
			}
		}
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "🔒 This channel has been locked.")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
	}

	_, err = s.ChannelMessageSend(config.BotLogID, "🔒 "+misc.ChMention(cha)+" was locked by "+m.Author.Username)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
	}
}

// Unlocks a specific channel and is spoiler-channel sensitive
func unlockCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		def              int
		roleID           string
		roleTempPosition int
		spoilerRole 	 = false
	)

	// Pulls info on the channel the message is in
	cha, err := s.Channel(m.ChannelID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// If channel is already unlocked then notify user
	for index, permission := range cha.PermissionOverwrites {
		if permission.ID == config.ServerID {
			if cha.PermissionOverwrites[index].Deny != discordgo.PermissionSendMessages {
				_, err = s.ChannelMessageSend(m.ChannelID, "🔓 Channel already unlocked.")
				if err != nil {
					_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
					if err != nil {
						return
					}
					return
				}
				return
			}
		}
	}

	// Sets permission variable to be neutral for send messages
	def &= ^discordgo.PermissionSendMessages

	// Error if lock used in moderator category
	if cha.ParentID == "360994750796529665" {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Cannot lock a mod channel due to permission reasons.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fetches info on server roles from the server and puts it in deb
	deb, err := s.GuildRoles(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Updates opt-in-under and opt-in-above position
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == config.OptInUnder {
			misc.OptinUnderPosition = deb[i].Position
		} else if deb[i].Name == config.OptInAbove {
			misc.OptinAbovePosition = deb[i].Position
		} else if deb[i].Name == cha.Name {
			roleID = deb[i].ID
			roleTempPosition = deb[i].Position
		}
	}

	// Checks if the channel being locked is between the opt-ins
	for i := 0; i < len(deb); i++ {
		if roleTempPosition < misc.OptinUnderPosition &&
			roleTempPosition > misc.OptinAbovePosition {
			spoilerRole = true
			break
		}
	}

	if spoilerRole {
		// Adds send permissions only to the channel role if it's a spoiler channel
		err = s.ChannelPermissionSet(m.ChannelID, roleID, "role", misc.SpoilerPerms, 0)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}
	} else {
		// Adds send permission from @everyone
		err = s.ChannelPermissionSet(m.ChannelID, config.ServerID, "role", def, 0)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
			return
		}
	}
	_, err = s.ChannelMessageSend(m.ChannelID, "🔓 This channel has been unlocked.")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
	}
	_, err = s.ChannelMessageSend(config.BotLogID, "🔓 "+misc.ChMention(cha)+" was unlocked by "+m.Author.Username)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
	}
}

func init() {
	add(&command{
		execute: lockCommand,
		trigger: "lock",
		desc:    "Locks a channel.",
		elevated: true,
		category: "channel",
	})
	add(&command{
		execute: unlockCommand,
		trigger: "unlock",
		desc:    "Unlocks a channel.",
		elevated: true,
		category: "channel",
	})
}