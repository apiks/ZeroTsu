package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Locks a specific channel and is spoiler-channel sensitive
func lockCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		roleExists       		bool
		roleID           		string
		airingID		 		string
		originalRolePerms		*discordgo.PermissionOverwrite
		originalAiringPerms		*discordgo.PermissionOverwrite
		originalEveryonePerms	*discordgo.PermissionOverwrite
	)

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Pulls info on the channel the message is in
	cha, err := s.Channel(m.ChannelID)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Fetches info on server roles from the server and puts it in roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Checks if the channel has an associated role and updates airing role location if it exists
	for _, role := range roles {
		if strings.ToLower(role.Name) == strings.ToLower(cha.Name) &&
			role.ID != m.GuildID {
			for roleID := range misc.GuildMap[m.GuildID].SpoilerMap {
				if role.ID == roleID {
					roleID = role.ID
					break
				}
			}
		}
		if strings.ToLower(role.Name) == "airing" {
			airingID = role.ID
		}
	}

	// Saves the original role and airing perms if they exists
	for _, perm := range cha.PermissionOverwrites {
		if perm.ID == roleID && roleID != "" {
			originalRolePerms = perm
		}
		if perm.ID == airingID && airingID != "" {
			originalAiringPerms = perm
		}
		if perm.ID == m.GuildID {
			originalEveryonePerms = perm
		}
	}

	// Removes send permissions from everyone, channel role and airing role
	if originalRolePerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, roleID, "role", originalRolePerms.Allow & ^discordgo.PermissionSendMessages, originalRolePerms.Deny | discordgo.PermissionSendMessages)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
	}
	if originalAiringPerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, airingID, "role", originalAiringPerms.Allow & ^discordgo.PermissionSendMessages, originalAiringPerms.Deny | discordgo.PermissionSendMessages)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
	}
	if originalEveryonePerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, m.GuildID, "role", originalEveryonePerms.Allow & ^discordgo.PermissionSendMessages, originalEveryonePerms.Deny | discordgo.PermissionSendMessages)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
	}

	// Adds mod role overwrites if they don't exist
	misc.MapMutex.Lock()
	for _, perm := range cha.PermissionOverwrites {
		for _, modRole := range misc.GuildMap[m.GuildID].GuildConfig.CommandRoles {
			if perm.ID == modRole.ID {
				roleExists = true
				break
			}
		}
		if roleExists {
			break
		}
	}
	if !roleExists {
		for _, modRole := range misc.GuildMap[m.GuildID].GuildConfig.CommandRoles {
			err = s.ChannelPermissionSet(m.ChannelID, modRole.ID, "role", discordgo.PermissionAll, 0)
			if err != nil {
				_, err = s.ChannelMessageSend(guildBotLog, err.Error() + "\n" + misc.ErrorLocation(err))
				if err != nil {
					return
				}
				return
			}
		}
	}
	misc.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, "🔒 This channel has been locked.")
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(guildBotLog, "🔒 "+misc.ChMention(cha)+" was locked by "+m.Author.Username)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Unlocks a specific channel and is spoiler-channel sensitive
func unlockCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		roleExists				bool
		roleID					string
		airingID				string
		originalRolePerms		*discordgo.PermissionOverwrite
		originalAiringPerms		*discordgo.PermissionOverwrite
		originalEveryonePerms	*discordgo.PermissionOverwrite
	)

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Pulls info on the channel the message is in
	cha, err := s.Channel(m.ChannelID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Fetches info on server roles from the server and puts it in roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Checks if the channel has an associated role and updates airing role location if it exists
	for _, role := range roles {
		if strings.ToLower(role.Name) == strings.ToLower(cha.Name) &&
			role.ID != m.GuildID {
			for rolID := range misc.GuildMap[m.GuildID].SpoilerMap {
				if role.ID == rolID {
					roleID = role.ID
					break
				}
			}
		}
		if strings.ToLower(role.Name) == "airing" {
			airingID = role.ID
		}
	}

	// Saves the original role and airing perms if they exists
	for _, perm := range cha.PermissionOverwrites {
		if perm.ID == roleID && roleID != "" {
			originalRolePerms = perm
		}
		if perm.ID == airingID && airingID != "" {
			originalAiringPerms = perm
		}
		if perm.ID == m.GuildID {
			originalEveryonePerms = perm
		}
	}

	// Adds send permissions to the channel role and airing if it's a spoiler channel
	if originalRolePerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, roleID, "role", originalRolePerms.Allow | discordgo.PermissionSendMessages, originalRolePerms.Deny & ^discordgo.PermissionSendMessages)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
	}
	if originalAiringPerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, airingID, "role", originalAiringPerms.Allow | discordgo.PermissionSendMessages, originalAiringPerms.Deny & ^discordgo.PermissionSendMessages)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
	}
	if originalEveryonePerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, m.GuildID, "role", originalEveryonePerms.Allow, originalEveryonePerms.Deny & ^discordgo.PermissionSendMessages)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
	}

	// Adds mod role overwrites if they don't exist
	misc.MapMutex.Lock()
	for _, perm := range cha.PermissionOverwrites {
		for _, modRole := range misc.GuildMap[m.GuildID].GuildConfig.CommandRoles {
			if perm.ID == modRole.ID {
				roleExists = true
				break
			}
		}
		if roleExists {
			break
		}
	}
	if !roleExists {
		for _, modRole := range misc.GuildMap[m.GuildID].GuildConfig.CommandRoles {
			err = s.ChannelPermissionSet(m.ChannelID, modRole.ID, "role", discordgo.PermissionAll, 0)
			if err != nil {
				misc.MapMutex.Unlock()
				misc.CommandErrorHandler(s, m, err, guildBotLog)
				return
			}
		}
	}
	misc.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, "🔓 This channel has been unlocked.")
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	_, err = s.ChannelMessageSend(guildBotLog, "🔓 "+misc.ChMention(cha)+" was unlocked by "+m.Author.Username)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
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