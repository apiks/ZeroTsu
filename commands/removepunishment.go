package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Removes a warning log entry via index from memberInfo entry
func removeWarningCommand(s *discordgo.Session, m *discordgo.Message) {

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"removewarning [@user, userID, or username#discrim] [warning index]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := misc.GetUserID(m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Checks if user is in memberInfo
	misc.MapMutex.Lock()
	if misc.GuildMap[m.GuildID].MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in internal database. Cannot remove nonexisting warning.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid warning index.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	misc.MapMutex.Lock()
	if index > len(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid warning index.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	// Removes warning from map and sets punishment
	misc.MapMutex.Lock()
	punishment := misc.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings[index]
	for timestampIndex, timestamp := range misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps {
		if strings.ToLower(timestamp.Punishment) == strings.ToLower(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings[index]) {
			misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[:timestampIndex], misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[timestampIndex+1:]...)
			break
		}
	}
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings[:index], misc.GuildMap[m.GuildID].MemberInfoMap[userID].Warnings[index+1:]...)

	// Writes new map to storage
	misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	misc.MapMutex.Unlock()

	err = removePunishmentEmbed(s, m, punishment)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Removes a kick log entry via index from memberInfo entry
func removeKickCommand(s *discordgo.Session, m *discordgo.Message) {

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and index.) Else prints error message
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"removekick [@user, userID, or username#discrim] [kick index]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := misc.GetUserID(m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Checks if user is in memberInfo
	misc.MapMutex.Lock()
	if misc.GuildMap[m.GuildID].MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in internal database. Cannot remove nonexisting kick.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid kick index.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	misc.MapMutex.Lock()
	if index > len(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid kick index.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	// Removes kick from map and sets punishment
	misc.MapMutex.Lock()
	punishment := misc.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks[index]
	for timestampIndex, timestamp := range misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps {
		if strings.ToLower(timestamp.Punishment) == strings.ToLower(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks[index]) {
			misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[:timestampIndex], misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[timestampIndex+1:]...)
			break
		}
	}
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks[:index], misc.GuildMap[m.GuildID].MemberInfoMap[userID].Kicks[index+1:]...)

	// Writes new map to storage
	misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	misc.MapMutex.Unlock()

	err = removePunishmentEmbed(s, m, punishment)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Removes a ban log entry via index from memberInfo entry
func removeBanCommand(s *discordgo.Session, m *discordgo.Message) {

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and index. Else prints error message
	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"removeban [@user, userID, or username#discrim] [ban index]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := misc.GetUserID(m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Checks if user is in memberInfo
	misc.MapMutex.Lock()
	if misc.GuildMap[m.GuildID].MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in internal database. Cannot remove nonexisting ban.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid ban index.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	misc.MapMutex.Lock()
	if index > len(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Bans) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid ban index.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	// Removes ban from map and sets punishment
	misc.MapMutex.Lock()
	punishment := misc.GuildMap[m.GuildID].MemberInfoMap[userID].Bans[index]
	for timestampIndex, timestamp := range misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps {
		if strings.ToLower(timestamp.Punishment) == strings.ToLower(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Bans[index]) {
			misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[:timestampIndex], misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[timestampIndex+1:]...)
			break
		}
	}
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Bans = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Bans[:index], misc.GuildMap[m.GuildID].MemberInfoMap[userID].Bans[index+1:]...)

	// Writes new map to storage
	misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	misc.MapMutex.Unlock()

	err = removePunishmentEmbed(s, m, punishment)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Removes a mute log entry via index from memberInfo entry
func removeMuteCommand(s *discordgo.Session, m *discordgo.Message) {

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and index. Else prints error message
	if len(commandStrings) < 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"removemute [@user, userID, or username#discrim] [mute index]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := misc.GetUserID(m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Checks if user is in memberInfo
	misc.MapMutex.Lock()
	if misc.GuildMap[m.GuildID].MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in internal database. Cannot remove nonexisting mute.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid mute index.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	misc.MapMutex.Lock()
	if index > len(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid mute index.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	// Removes mute from map and sets punishment
	misc.MapMutex.Lock()
	punishment := misc.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes[index]
	for timestampIndex, timestamp := range misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps {
		if strings.ToLower(timestamp.Punishment) == strings.ToLower(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes[index]) {
			misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[:timestampIndex], misc.GuildMap[m.GuildID].MemberInfoMap[userID].Timestamps[timestampIndex+1:]...)
			break
		}
	}
	misc.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes = append(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes[:index], misc.GuildMap[m.GuildID].MemberInfoMap[userID].Mutes[index+1:]...)

	// Writes new map to storage
	misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	misc.MapMutex.Unlock()

	err = removePunishmentEmbed(s, m, punishment)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func removePunishmentEmbed(s *discordgo.Session, m *discordgo.Message, punishment string) error {

	var embedMess discordgo.MessageEmbed

	// Sets punishment embed color
	embedMess.Color = 16758465
	embedMess.Title = fmt.Sprintf("Successfuly removed punishment: _%v_", punishment)

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	return err
}

func init() {
	add(&command{
		execute:  removeWarningCommand,
		trigger:  "removewarning",
		desc:     "Removes a user warning by number",
		elevated: true,
		category: "moderation",
	})
	add(&command{
		execute:  removeKickCommand,
		trigger:  "removekick",
		desc:     "Removes a user kick by number",
		elevated: true,
		category: "moderation",
	})
	add(&command{
		execute:  removeBanCommand,
		trigger:  "removeban",
		desc:     "Removes a user ban by number",
		elevated: true,
		category: "moderation",
	})
	add(&command{
		execute:  removeMuteCommand,
		trigger:  "removemute",
		desc:     "Removes a user mute by number",
		elevated: true,
		category: "moderation",
	})
}
