package commands

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Removes a warning log entry via index from memberInfo entry
func removeWarningCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Checks if there's enough parameters
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"removewarning [@user, userID, or username#discrim] [warning index]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := functionality.GetUserID(m, commandStrings)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Checks if user is in memberInfo
	functionality.MapMutex.Lock()
	if functionality.GuildMap[m.GuildID].MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in internal database. Cannot remove nonexisting warning.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}
	functionality.MapMutex.Unlock()

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid warning index.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	functionality.MapMutex.Lock()
	if index > len(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid warning index.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}
	functionality.MapMutex.Unlock()

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	// Removes warning from map and sets punishment
	functionality.MapMutex.Lock()
	punishment := functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings[index]
	for timestampIndex, timestamp := range functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps {
		if strings.ToLower(timestamp.Punishment) == strings.ToLower(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings[index]) {
			functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[:timestampIndex], functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[timestampIndex+1:]...)
			break
		}
	}
	functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings[:index], functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings[index+1:]...)

	// Writes new map to storage
	_ = functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	functionality.MapMutex.Unlock()

	err = functionality.RemovePunishmentEmbed(s, m, punishment)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a kick log entry via index from memberInfo entry
func removeKickCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Checks if there's enough parameters (command, user and index.) Else prints error message
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"removekick [@user, userID, or username#discrim] [kick index]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := functionality.GetUserID(m, commandStrings)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Checks if user is in memberInfo
	functionality.MapMutex.Lock()
	if functionality.GuildMap[m.GuildID].MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in internal database. Cannot remove nonexisting kick.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}
	functionality.MapMutex.Unlock()

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid kick index.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	functionality.MapMutex.Lock()
	if index > len(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid kick index.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}
	functionality.MapMutex.Unlock()

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	// Removes kick from map and sets punishment
	functionality.MapMutex.Lock()
	punishment := functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks[index]
	for timestampIndex, timestamp := range functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps {
		if strings.ToLower(timestamp.Punishment) == strings.ToLower(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks[index]) {
			functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[:timestampIndex], functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[timestampIndex+1:]...)
			break
		}
	}
	functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks[:index], functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks[index+1:]...)

	// Writes new map to storage
	_ = functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	functionality.MapMutex.Unlock()

	err = functionality.RemovePunishmentEmbed(s, m, punishment)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a ban log entry via index from memberInfo entry
func removeBanCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Checks if there's enough parameters (command, user and index. Else prints error message
	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"removeban [@user, userID, or username#discrim] [ban index]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := functionality.GetUserID(m, commandStrings)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Checks if user is in memberInfo
	functionality.MapMutex.Lock()
	if functionality.GuildMap[m.GuildID].MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in internal database. Cannot remove nonexisting ban.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}
	functionality.MapMutex.Unlock()

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid ban index.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	functionality.MapMutex.Lock()
	if index > len(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Bans) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid ban index.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}
	functionality.MapMutex.Unlock()

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	// Removes ban from map and sets punishment
	functionality.MapMutex.Lock()
	punishment := functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Bans[index]
	for timestampIndex, timestamp := range functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps {
		if strings.ToLower(timestamp.Punishment) == strings.ToLower(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Bans[index]) {
			functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[:timestampIndex], functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[timestampIndex+1:]...)
			break
		}
	}
	functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Bans = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Bans[:index], functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Bans[index+1:]...)

	// Writes new map to storage
	_ = functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	functionality.MapMutex.Unlock()

	err = functionality.RemovePunishmentEmbed(s, m, punishment)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a mute log entry via index from memberInfo entry
func removeMuteCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Checks if there's enough parameters (command, user and index. Else prints error message
	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"removemute [@user, userID, or username#discrim] [mute index]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := functionality.GetUserID(m, commandStrings)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Checks if user is in memberInfo
	functionality.MapMutex.Lock()
	if functionality.GuildMap[m.GuildID].MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in internal database. Cannot remove nonexisting mute.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}
	functionality.MapMutex.Unlock()

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid mute index.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	functionality.MapMutex.Lock()
	if index > len(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid mute index.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}
	functionality.MapMutex.Unlock()

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	// Removes mute from map and sets punishment
	functionality.MapMutex.Lock()
	punishment := functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes[index]
	for timestampIndex, timestamp := range functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps {
		if strings.ToLower(timestamp.Punishment) == strings.ToLower(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes[index]) {
			functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[:timestampIndex], functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[timestampIndex+1:]...)
			break
		}
	}
	functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes = append(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes[:index], functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes[index+1:]...)

	// Writes new map to storage
	_ = functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	functionality.MapMutex.Unlock()

	err = functionality.RemovePunishmentEmbed(s, m, punishment)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    removeWarningCommand,
		Trigger:    "removewarning",
		Desc:       "Removes a user warning by number",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
	functionality.Add(&functionality.Command{
		Execute:    removeKickCommand,
		Trigger:    "removekick",
		Desc:       "Removes a user kick by number",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
	functionality.Add(&functionality.Command{
		Execute:    removeBanCommand,
		Trigger:    "removeban",
		Desc:       "Removes a user ban by number",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
	functionality.Add(&functionality.Command{
		Execute:    removeMuteCommand,
		Trigger:    "removemute",
		Desc:       "Removes a user mute by number",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
