package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Removes a warning log entry via index from memberInfo entry
func removeWarningCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Checks if there's enough parameters
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"removewarning [@user, userID, or username#discrim] [warning index]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := common.GetUserID(m, commandStrings)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid warning index.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks and fetches user
	mem := db.GetGuildMember(m.GuildID, userID)
	if mem.GetID() == "" {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, fmt.Errorf("Error: User does not exist in the internal database. Cannot remove nonexisting warning."))
		return
	}

	if index > len(mem.GetWarnings()) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid warning index.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	// Removes warning from map and sets punishment
	punishment := mem.GetWarnings()[index]
	for i, timestamp := range mem.GetTimestamps() {
		if strings.ToLower(timestamp.GetPunishment()) == strings.ToLower(mem.GetWarnings()[index]) {
			mem = mem.RemoveFromTimestamps(i)
			break
		}
	}
	mem = mem.RemoveFromWarnings(index)

	// Write
	db.SetGuildMember(m.GuildID, mem)

	err = embeds.PunishmentRemoval(s, m, "warning", punishment)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a mute log entry via index from memberInfo entry
func removeMuteCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Checks if there's enough parameters (command, user and index. Else prints error message
	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"removemute [@user, userID, or username#discrim] [mute index]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := common.GetUserID(m, commandStrings)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid mute index.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks and fetches user
	mem := db.GetGuildMember(m.GuildID, userID)
	if mem.GetID() == "" {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, fmt.Errorf("Error: User does not exist in the internal database. Cannot remove nonexisting mute."))
		return
	}

	if index > len(mem.GetMutes()) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid mute index.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	// Removes mute from map and sets punishment
	punishment := mem.GetMutes()[index]
	for i, timestamp := range mem.GetTimestamps() {
		if strings.ToLower(timestamp.GetPunishment()) == strings.ToLower(mem.GetMutes()[index]) {
			mem = mem.RemoveFromTimestamps(i)
			break
		}
	}
	mem = mem.RemoveFromMutes(index)

	// Write
	db.SetGuildMember(m.GuildID, mem)

	err = embeds.PunishmentRemoval(s, m, "mute", punishment)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a kick log entry via index from memberInfo entry
func removeKickCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)
	cmdStrs := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Checks if there's enough parameters (command, user and index.) Else prints error message
	if len(cmdStrs) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"removekick [@user, userID, or username#discrim] [kick index]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of cmdStrs, else print error
	userID, err := common.GetUserID(m, cmdStrs)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Index checks
	index, err := strconv.Atoi(cmdStrs[2])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid kick index.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks and fetches user
	mem := db.GetGuildMember(m.GuildID, userID)
	if mem.GetID() == "" {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, fmt.Errorf("Error: User does not exist in the internal database. Cannot remove nonexisting kick."))
		return
	}

	if index > len(mem.GetKicks()) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid kick index.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	// Removes kick from map and sets punishment
	punishment := mem.GetKicks()[index]
	for i, timestamp := range mem.GetTimestamps() {
		if strings.ToLower(timestamp.GetPunishment()) == strings.ToLower(mem.GetKicks()[index]) {
			mem = mem.RemoveFromTimestamps(i)
			break
		}
	}
	mem = mem.RemoveFromKicks(index)

	// Write
	db.SetGuildMember(m.GuildID, mem)

	err = embeds.PunishmentRemoval(s, m, "kick", punishment)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a ban log entry via index from memberInfo entry
func removeBanCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Checks if there's enough parameters (command, user and index. Else prints error message
	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"removeban [@user, userID, or username#discrim] [ban index]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := common.GetUserID(m, commandStrings)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid ban index.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks and fetches user
	mem := db.GetGuildMember(m.GuildID, userID)
	if mem.GetID() == "" {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, fmt.Errorf("Error: User does not exist in the internal database. Cannot remove nonexisting ban."))
		return
	}

	if index > len(mem.GetBans()) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid ban index.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	// Removes ban from map and sets punishment
	punishment := mem.GetBans()[index]
	for i, timestamp := range mem.GetTimestamps() {
		if strings.ToLower(timestamp.GetPunishment()) == strings.ToLower(mem.GetBans()[index]) {
			mem = mem.RemoveFromTimestamps(i)
			break
		}
	}
	mem = mem.RemoveFromBans(index)

	// Write
	db.SetGuildMember(m.GuildID, mem)

	err = embeds.PunishmentRemoval(s, m, "ban", punishment)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func init() {
	Add(&Command{
		Execute:    removeWarningCommand,
		Trigger:    "removewarning",
		Desc:       "Removes a user warning by number",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
	Add(&Command{
		Execute:    removeKickCommand,
		Trigger:    "removekick",
		Desc:       "Removes a user kick by number",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
	Add(&Command{
		Execute:    removeBanCommand,
		Trigger:    "removeban",
		Desc:       "Removes a user ban by number",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
	Add(&Command{
		Execute:    removeMuteCommand,
		Trigger:    "removemute",
		Desc:       "Removes a user mute by number",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
