package commands

import (
	"strings"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
	"github.com/r-anime/ZeroTsu/config"
)

// Removes a warning log entry via index from memberInfo entry
func removeWarningCommand(s *discordgo.Session, m *discordgo.Message) {

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "removewarning [@user or userID] [warning index]")
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

	// Pulls info on user
	userMem, err := s.User(userID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Checks if user is in memberInfo
	if misc.MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in memberInfo.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {

				return
			}
			return
		}
		return
	}

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
	if index > len(misc.MemberInfoMap[userID].Warnings) || index < 0 {
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

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success. Removed warning `" + misc.MemberInfoMap[userID].Warnings[index] +
		"` from " + userMem.Username + "#" + userMem.Discriminator)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
	}

	// Removes warning from map
	misc.MapMutex.Lock()
	misc.MemberInfoMap[userID].Warnings = append(misc.MemberInfoMap[userID].Warnings[:index], misc.MemberInfoMap[userID].Warnings[index+1:]...)
	misc.MapMutex.Unlock()

	// Writes new map to storage
	misc.MemberInfoWrite(misc.MemberInfoMap)
}

// Removes a kick log entry via index from memberInfo entry
func removeKickCommand(s *discordgo.Session, m *discordgo.Message) {

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and index.) Else prints error message
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "removekick [@user or userID] [kick index]")
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

	// Pulls info on user
	userMem, err := s.User(userID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Checks if user is in memberInfo
	if misc.MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in memberInfo.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

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
	if index > len(misc.MemberInfoMap[userID].Warnings) || index < 0 {

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

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success. Removed kick `" + misc.MemberInfoMap[userID].Kicks[index] +
		"` from " + userMem.Username + "#" + userMem.Discriminator)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
	}

	// Removes warning from map
	misc.MapMutex.Lock()
	misc.MemberInfoMap[userID].Kicks = append(misc.MemberInfoMap[userID].Kicks[:index], misc.MemberInfoMap[userID].Kicks[index+1:]...)
	misc.MapMutex.Unlock()

	// Writes new map to storage
	misc.MemberInfoWrite(misc.MemberInfoMap)
}

// Removes a ban log entry via index from memberInfo entry
func removeBanCommand(s *discordgo.Session, m *discordgo.Message) {

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and index. Else prints error message
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "removeban [@user or userID] [ban index]")
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

	// Pulls info on user
	userMem, err := s.User(userID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Checks if user is in memberInfo
	if misc.MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in memberInfo.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

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
	if index > len(misc.MemberInfoMap[userID].Warnings) || index < 0 {
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

	// Fixes index for future use if it's 0
	if index != 0 {
		index = index - 1
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success. Removed ban `" + misc.MemberInfoMap[userID].Bans[index] +
		"` from " + userMem.Username + "#" + userMem.Discriminator)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
	}

	// Removes warning from map
	misc.MapMutex.Lock()
	misc.MemberInfoMap[userID].Bans = append(misc.MemberInfoMap[userID].Bans[:index], misc.MemberInfoMap[userID].Bans[index+1:]...)
	misc.MapMutex.Unlock()

	// Writes new map to storage
	misc.MemberInfoWrite(misc.MemberInfoMap)
}

//func init() {
//	add(&command{
//		execute:  removeWarningCommand,
//		trigger:  "removewarning",
//		desc:     "Removes a user warning whois text",
//		elevated: true,
//		category: "punishment",
//	})
//	add(&command{
//		execute:  removeKickCommand,
//		trigger:  "removekick",
//		desc:     "Removes a user kick whois text",
//		elevated: true,
//		category: "punishment",
//	})
//	add(&command{
//		execute:  removeBanCommand,
//		trigger:  "removeban",
//		desc:     "Removes a user ban whois text",
//		elevated: true,
//		category: "punishment",
//	})
//}