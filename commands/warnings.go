package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
	"github.com/r-anime/ZeroTsu/config"
)

// Adds a warning to a specific user in memberInfo.json without telling them
func addWarningCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		warning string
	)

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 3)

	if len(commandStrings) != 3 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"addwarning [@user or userID] [warning]`")
		if err != nil {

			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {

				return
			}
			return
		}
		return
	}

	userID := misc.GetUserID(s, m, commandStrings)
	if userID == "" {
		return
	}

	warning = commandStrings[2]

	// If memberInfo.json file is empty or user is not there, print error
	if misc.MemberInfoMap == nil || misc.MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo.")
		if err != nil {

			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {

				return
			}
			return
		}
		return
	}

	// Appends warning to user in memberInfo
	misc.MapMutex.Lock()
	misc.MemberInfoMap[userID].Warnings = append(misc.MemberInfoMap[userID].Warnings, warning)
	misc.MapMutex.Unlock()

	// Writes to memberInfo.json
	misc.MemberInfoWrite(misc.MemberInfoMap)

	// Pulls info on user
	userMem, err := s.User(userID)
	if err != nil {

		misc.CommandErrorHandler(s, m, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, userMem.Username + "#" + userMem.Discriminator + " had warning added: " + "`" + warning + "`")
	if err != nil {

		_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
		if err != nil {

			return
		}
		return
	}
}

// Issues a warning to a specific user in memberInfo.json wand tells them
func issueWarningCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		warning string
	)

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 3)

	if len(commandStrings) != 3 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"issuewarning [@user or userID] [warning]`")
		if err != nil {

			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {

				return
			}
			return
		}
		return
	}

	userID := misc.GetUserID(s, m, commandStrings)
	if userID == "" {
		return
	}

	warning = commandStrings[2]

	// If memberInfo.json file is empty or user is not there, print error
	if misc.MemberInfoMap == nil || misc.MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo.")
		if err != nil {

			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {

				return
			}
			return
		}
		return
	}

	// Appends warning to user in memberInfo
	misc.MapMutex.Lock()
	misc.MemberInfoMap[userID].Warnings = append(misc.MemberInfoMap[userID].Warnings, warning)
	misc.MapMutex.Unlock()

	// Writes to memberInfo.json
	misc.MemberInfoWrite(misc.MemberInfoMap)

	// Pulls info on user
	userMem, err := s.User(userID)
	if err != nil {

		misc.CommandErrorHandler(s, m, err)
		return
	}

	//Pulls the guild Name
	guild, err := s.Guild(config.ServerID)
	if err != nil {

		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Sends message in DMs that they have been banned if able
	dm, err := s.UserChannelCreate(userID)
	if err != nil {

		return
	}
	_, _ = s.ChannelMessageSend(dm.ID, "You have been warned on " + guild.Name + ":\n`" + warning + "`")

	// Sends mod success message
	_, err = s.ChannelMessageSend(m.ChannelID, userMem.Username + "#" + userMem.Discriminator + " was warned with: " + "`" + warning + "`")
	if err != nil {

		_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
		if err != nil {

			return
		}
		return
	}
}

//func init() {
//	add(&command{
//		execute:  addWarningCommand,
//		trigger:  "addwarning",
//		desc:     "Adds a warning to a user without telling them",
//		elevated: true,
//	})
//	add(&command{
//		execute:  issueWarningCommand,
//		trigger:  "issuewarning",
//		desc:     "Issues a warning to a user and tells them",
//		elevated: true,
//	})
//}