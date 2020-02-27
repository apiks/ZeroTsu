package commands

import (
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
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

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildSpoilerMap := db.GetGuildSpoilerMap(m.GuildID)

	// Pulls info on the channel the message is in
	cha, err := s.Channel(m.ChannelID)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Fetches info on server roles from the server and puts it in roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Checks if the channel has an associated role and updates airing role location if it exists
	for _, role := range roles {
		if strings.ToLower(role.Name) == strings.ToLower(cha.Name) &&
			role.ID != m.GuildID {
			for roleID := range guildSpoilerMap {
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
		err = s.ChannelPermissionSet(m.ChannelID, roleID, "role", originalRolePerms.Allow & ^discordgo.PermissionSendMessages, originalRolePerms.Deny|discordgo.PermissionSendMessages)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
	if originalAiringPerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, airingID, "role", originalAiringPerms.Allow & ^discordgo.PermissionSendMessages, originalAiringPerms.Deny|discordgo.PermissionSendMessages)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
	if originalEveryonePerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, m.GuildID, "role", originalEveryonePerms.Allow & ^discordgo.PermissionSendMessages, originalEveryonePerms.Deny|discordgo.PermissionSendMessages)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	} else {
		err = s.ChannelPermissionSet(m.ChannelID, m.GuildID, "role", 0, discordgo.PermissionSendMessages)
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}

	// Adds mod role overwrites if they don't exist
	for _, perm := range cha.PermissionOverwrites {
		for _, modRole := range guildSettings.GetCommandRoles() {
			if perm.ID == modRole.GetID() {
				roleExists = true
				break
			}
		}
		if roleExists {
			break
		}
	}
	if !roleExists {
		for _, modRole := range guildSettings.GetCommandRoles() {
			err = s.ChannelPermissionSet(m.ChannelID, modRole.GetID(), "role", discordgo.PermissionSendMessages, 0)
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "🔒 This channel has been locked.")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}

	if guildSettings.BotLog == (entities.Cha{}) {
		return
	}
	if guildSettings.BotLog.GetID() == "" {
		return
	}
	_, _ = s.ChannelMessageSend(guildSettings.BotLog.GetID(), "🔒 "+common.ChMention(cha)+" was locked by "+m.Author.Username)
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

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildSpoilerMap := db.GetGuildSpoilerMap(m.GuildID)

	// Pulls info on the channel the message is in
	cha, err := s.Channel(m.ChannelID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Fetches info on server roles from the server and puts it in roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Checks if the channel has an associated role and updates airing role location if it exists
	for _, role := range roles {
		if strings.ToLower(role.Name) == strings.ToLower(cha.Name) &&
			role.ID != m.GuildID {
			for roleID := range guildSpoilerMap {
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

	// Adds send permissions to the channel role and airing if it's a spoiler channel
	if originalRolePerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, roleID, "role", originalRolePerms.Allow|discordgo.PermissionSendMessages, originalRolePerms.Deny & ^discordgo.PermissionSendMessages)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}
	if originalAiringPerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, airingID, "role", originalAiringPerms.Allow|discordgo.PermissionSendMessages, originalAiringPerms.Deny & ^discordgo.PermissionSendMessages)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}
	if originalEveryonePerms != nil {
		err = s.ChannelPermissionSet(m.ChannelID, m.GuildID, "role", originalEveryonePerms.Allow, originalEveryonePerms.Deny & ^discordgo.PermissionSendMessages)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	// Adds mod role overwrites if they don't exist
	for _, perm := range cha.PermissionOverwrites {
		for _, modRole := range guildSettings.GetCommandRoles() {
			if perm.ID == modRole.GetID() {
				roleExists = true
				break
			}
		}
		if roleExists {
			break
		}
	}
	if !roleExists {
		for _, modRole := range guildSettings.GetCommandRoles() {
			err = s.ChannelPermissionSet(m.ChannelID, modRole.GetID(), "role", discordgo.PermissionSendMessages, 0)
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "🔓 This channel has been unlocked.")
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	if guildSettings.BotLog == (entities.Cha{}) {
		return
	}
	if guildSettings.BotLog.GetID() == "" {
		return
	}
	_, _ = s.ChannelMessageSend(guildSettings.BotLog.GetID(), "🔓 "+common.ChMention(cha)+" was unlocked by "+m.Author.Username)
}

func init() {
	Add(&Command{
		Execute:    lockCommand,
		Trigger:    "lock",
		Aliases:    []string{"lockchannel", "channellock"},
		Desc:       "Locks a channel",
		Permission: functionality.Mod,
		Module:     "channel",
	})
	Add(&Command{
		Execute:    unlockCommand,
		Trigger:    "unlock",
		Aliases:    []string{"unlockchannel", "channelunlock"},
		Desc:       "Unlocks a channel",
		Permission: functionality.Mod,
		Module:     "channel",
	})
}
