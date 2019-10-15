package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Locks a specific channel and is spoiler-channel sensitive
func lockCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		roleExists            bool
		roleID                string
		airingID              string
		originalRolePerms     *discordgo.PermissionOverwrite
		originalAiringPerms   *discordgo.PermissionOverwrite
		originalEveryonePerms *discordgo.PermissionOverwrite
	)

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	// Pulls info on the channel the message is in
	cha, err := s.Channel(m.ChannelID)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Fetches info on server roles from the server and puts it in roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Checks if the channel has an associated role and updates airing role location if it exists
	functionality.MapMutex.Lock()
	for _, role := range roles {
		if strings.ToLower(role.Name) == strings.ToLower(cha.Name) &&
			role.ID != m.GuildID {
			for roleID := range functionality.GuildMap[m.GuildID].SpoilerMap {
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
	functionality.MapMutex.Unlock()

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
		err = s.ChannelPermissionSet(m.ChannelID, roleID, "role", originalRolePerms.Allow & ^discordgo.PermissionSendMessages, originalRolePerms.Deny|discordgo.PermissionSendMessages)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
	if originalAiringPerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, airingID, "role", originalAiringPerms.Allow & ^discordgo.PermissionSendMessages, originalAiringPerms.Deny|discordgo.PermissionSendMessages)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
	if originalEveryonePerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, m.GuildID, "role", originalEveryonePerms.Allow & ^discordgo.PermissionSendMessages, originalEveryonePerms.Deny|discordgo.PermissionSendMessages)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
	} else {
		err = s.ChannelPermissionSet(m.ChannelID, m.GuildID, "role", 0, discordgo.PermissionSendMessages)
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
	}

	// Adds mod role overwrites if they don't exist
	for _, perm := range cha.PermissionOverwrites {
		for _, modRole := range guildSettings.CommandRoles {
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
		for _, modRole := range guildSettings.CommandRoles {
			err = s.ChannelPermissionSet(m.ChannelID, modRole.ID, "role", discordgo.PermissionSendMessages, 0)
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "🔒 This channel has been locked.")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}

	if guildSettings.BotLog == nil {
		return
	}
	if guildSettings.BotLog.ID == "" {
		return
	}
	_, _ = s.ChannelMessageSend(guildSettings.BotLog.ID, "🔒 "+functionality.ChMention(cha)+" was locked by "+m.Author.Username)
}

// Unlocks a specific channel and is spoiler-channel sensitive
func unlockCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		roleExists            bool
		roleID                string
		airingID              string
		originalRolePerms     *discordgo.PermissionOverwrite
		originalAiringPerms   *discordgo.PermissionOverwrite
		originalEveryonePerms *discordgo.PermissionOverwrite
	)

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	// Pulls info on the channel the message is in
	cha, err := s.Channel(m.ChannelID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Fetches info on server roles from the server and puts it in roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Checks if the channel has an associated role and updates airing role location if it exists
	functionality.MapMutex.Lock()
	for _, role := range roles {
		if strings.ToLower(role.Name) == strings.ToLower(cha.Name) &&
			role.ID != m.GuildID {
			for rolID := range functionality.GuildMap[m.GuildID].SpoilerMap {
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
	functionality.MapMutex.Unlock()

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
		err = s.ChannelPermissionSet(m.ChannelID, roleID, "role", originalRolePerms.Allow|discordgo.PermissionSendMessages, originalRolePerms.Deny & ^discordgo.PermissionSendMessages)
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}
	if originalAiringPerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, airingID, "role", originalAiringPerms.Allow|discordgo.PermissionSendMessages, originalAiringPerms.Deny & ^discordgo.PermissionSendMessages)
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}
	if originalEveryonePerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, m.GuildID, "role", originalEveryonePerms.Allow, originalEveryonePerms.Deny & ^discordgo.PermissionSendMessages)
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	// Adds mod role overwrites if they don't exist
	for _, perm := range cha.PermissionOverwrites {
		for _, modRole := range guildSettings.CommandRoles {
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
		for _, modRole := range guildSettings.CommandRoles {
			err = s.ChannelPermissionSet(m.ChannelID, modRole.ID, "role", discordgo.PermissionSendMessages, 0)
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "🔓 This channel has been unlocked.")
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	if guildSettings.BotLog == nil {
		return
	}
	if guildSettings.BotLog.ID == "" {
		return
	}
	_, _ = s.ChannelMessageSend(guildSettings.BotLog.ID, "🔓 "+functionality.ChMention(cha)+" was unlocked by "+m.Author.Username)
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    lockCommand,
		Trigger:    "lock",
		Aliases:    []string{"lockchannel", "channellock"},
		Desc:       "Locks a channel",
		Permission: functionality.Mod,
		Module:     "channel",
	})
	functionality.Add(&functionality.Command{
		Execute:    unlockCommand,
		Trigger:    "unlock",
		Aliases:    []string{"unlockchannel", "channelunlock"},
		Desc:       "Unlocks a channel",
		Permission: functionality.Mod,
		Module:     "channel",
	})
}
