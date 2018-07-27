package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
	"github.com/r-anime/ZeroTsu/config"
)

// Adds a warning to a specific user in memberInfo.json without telling them
func addWarningCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		warning string
		userID string
	)

	// Puts entire message in lowercase
	messageLowercase := strings.ToLower(m.Content)

	// Pulls the user and warning from strings after "addwarning"
	commandStrings := strings.SplitN(messageLowercase, " ", 3)

	// Pulls userID from 2nd parameter of commandStrings, else print error. Also pulls warning after it.
	if len(commandStrings) == 3 {

		userID := misc.GetUserID(s, m, commandStrings)
		if userID == "" {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid user. Please use `"+config.BotPrefix+"addwarning [@user or userID] [warning]` format.")
			if err != nil {

				return
			}
			return
		}
		warning = commandStrings[2]

	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"addwarning [@user or userID] [warning]`")
		if err != nil {

			return
		}
		return
	}

	// If memberInfo.json file is empty or user is not there, print error
	if misc.MemberInfoMap == nil || misc.MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo.")
		if err != nil {
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
		fmt.Println(err.Error())
	}

	_, err = s.ChannelMessageSend(m.ChannelID, userMem.Username + "#" + userMem.Discriminator + " had warning added: " + "`" + warning + "`")
	if err != nil {

		fmt.Println("Error:", err)
	}
}

// Issues a warning to a specific user in memberInfo.json wand tells them
func issueWarningCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		warning string
		userID string
	)

	// Puts entire message in lowercase
	messageLowercase := strings.ToLower(m.Content)

	// Pulls the user and warning from strings after "issuewarning"
	commandStrings := strings.SplitN(messageLowercase, " ", 3)

	// Pulls userID from 2nd parameter of commandStrings, else print error. Also pulls warning after it.
	if len(commandStrings) == 3 {

		userID := misc.GetUserID(s, m, commandStrings)
		if userID == "" {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid user. Please use `"+config.BotPrefix+"issuewarning [@user or userID] [warning]` format.")
			if err != nil {

				return
			}
			return
		}
		warning = commandStrings[2]

	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"issuewarning [@user or userID] [warning]`")
		if err != nil {

			return
		}
		return
	}

	// If memberInfo.json file is empty or user is not there, print error
	if misc.MemberInfoMap == nil || misc.MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo.")
		if err != nil {
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
		fmt.Println(err.Error())
	}

	//Pulls the guild Name
	guild, err := s.Guild(config.ServerID)
	if err != nil {

		fmt.Println("Error:", err)
	}

	// Sends message in DMs that they have been banned if able
	dm, err := s.UserChannelCreate(userID)
	if err != nil {

		return
	}
	_, err = s.ChannelMessageSend(dm.ID, "You have been warned on " + guild.Name + ":\n`" + warning + "`")

	// Sends mod success message
	_, err = s.ChannelMessageSend(m.ChannelID, userMem.Username + "#" + userMem.Discriminator + " was warned with: " + "`" + warning + "`")
	if err != nil {

		fmt.Println("Error:", err)
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