package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Removes a warning log entry via index from memberInfo entry
func removeWarningCommand(s *discordgo.Session, m *discordgo.Message) {

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "removewarning [@user, userID, or username#discrim] [warning index]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := misc.GetUserID(s, m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Checks if user is in memberInfo
	misc.MapMutex.Lock()
	if misc.MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in memberInfo.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
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
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	misc.MapMutex.Lock()
	if index > len(misc.MemberInfoMap[userID].Warnings) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid warning index.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
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
	punishment := misc.MemberInfoMap[userID].Warnings[index]
	misc.MemberInfoMap[userID].Warnings = append(misc.MemberInfoMap[userID].Warnings[:index], misc.MemberInfoMap[userID].Warnings[index+1:]...)
	misc.MapMutex.Unlock()

	// Writes new map to storage
	misc.MemberInfoWrite(misc.MemberInfoMap)

	err = removePunishmentEmbed(s, m, punishment)
	if err != nil {
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

// Removes a kick log entry via index from memberInfo entry
func removeKickCommand(s *discordgo.Session, m *discordgo.Message) {

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and index.) Else prints error message
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "removekick [@user, userID, or username#discrim] [kick index]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := misc.GetUserID(s, m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Checks if user is in memberInfo
	misc.MapMutex.Lock()
	if misc.MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in memberInfo.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
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
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	misc.MapMutex.Lock()
	if index > len(misc.MemberInfoMap[userID].Kicks) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid kick index.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
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
	punishment := misc.MemberInfoMap[userID].Kicks[index]
	misc.MemberInfoMap[userID].Kicks = append(misc.MemberInfoMap[userID].Kicks[:index], misc.MemberInfoMap[userID].Kicks[index+1:]...)
	misc.MapMutex.Unlock()

	// Writes new map to storage
	misc.MemberInfoWrite(misc.MemberInfoMap)

	err = removePunishmentEmbed(s, m, punishment)
	if err != nil {
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

// Removes a ban log entry via index from memberInfo entry
func removeBanCommand(s *discordgo.Session, m *discordgo.Message) {

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and index. Else prints error message
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "removeban [@user, userID, or username#discrim] [ban index]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := misc.GetUserID(s, m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Checks if user is in memberInfo
	misc.MapMutex.Lock()
	if misc.MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in memberInfo.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
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
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
	misc.MapMutex.Lock()
	if index > len(misc.MemberInfoMap[userID].Bans) || index < 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid ban index.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
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
	punishment := misc.MemberInfoMap[userID].Bans[index]
	misc.MemberInfoMap[userID].Bans = append(misc.MemberInfoMap[userID].Bans[:index], misc.MemberInfoMap[userID].Bans[index+1:]...)
	misc.MapMutex.Unlock()

	// Writes new map to storage
	misc.MemberInfoWrite(misc.MemberInfoMap)

	err = removePunishmentEmbed(s, m, punishment)
	if err != nil {
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

func removePunishmentEmbed(s *discordgo.Session, m *discordgo.Message, punishment string) error {

	var embedMess      discordgo.MessageEmbed

	// Sets punishment embed color
	embedMess.Color = 0x00ff00
	embedMess.Title = fmt.Sprintf("Successfuly removed punishment: _%v_", punishment)

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	return err
}

func init() {
	add(&command{
		execute:  removeWarningCommand,
		trigger:  "removewarning",
		desc:     "Removes a user warning by index.",
		elevated: true,
		category: "punishment",
	})
	add(&command{
		execute:  removeKickCommand,
		trigger:  "removekick",
		desc:     "Removes a user kick by index.",
		elevated: true,
		category: "punishment",
	})
	add(&command{
		execute:  removeBanCommand,
		trigger:  "removeban",
		desc:     "Removes a user ban by index.",
		elevated: true,
		category: "punishment",
	})
}