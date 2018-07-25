package commands

import (
	"strings"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

func removeWarningCommand(s *discordgo.Session, m *discordgo.Message) {

	// Puts entire message in lowercase
	messageLowercase := strings.ToLower(m.Content)

	//Separates every word in the message and puts it in a slice
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and index. Else prints error message
	if len(commandStrings) != 3 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Wrong amount of parameters.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID := misc.GetUserID(s, m, commandStrings)

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid warning index.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}
	if index > len(misc.MemberInfoMap[userID].Warnings) || index < 0 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid warning index.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Fixes index for future use if it's 0
	if index != 0 {

		index = index - 1
	}

	//Pulls info on user
	userMem, err := s.User(userID)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Checks if user is in memberInfo, giving error if not
	if misc.MemberInfoMap[userID] == nil {

		// Prints error
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in memberInfo.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Prints success
	_, err = s.ChannelMessageSend(m.ChannelID, "Success. Removed warning `" + misc.MemberInfoMap[userID].Warnings[index] +
		"` from " + userMem.Username + "#" + userMem.Discriminator)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Removes warning from map
	misc.MapMutex.Lock()
	misc.MemberInfoMap[userID].Warnings = append(misc.MemberInfoMap[userID].Warnings[:index], misc.MemberInfoMap[userID].Warnings[index+1:]...)
	misc.MapMutex.Unlock()

	// Writes new map to storage
	misc.MemberInfoWrite(misc.MemberInfoMap)
}

func removeKickCommand(s *discordgo.Session, m *discordgo.Message) {

	// Puts entire message in lowercase
	messageLowercase := strings.ToLower(m.Content)

	//Separates every word in the message and puts it in a slice
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and index. Else prints error message
	if len(commandStrings) != 3 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Wrong amount of parameters.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID := misc.GetUserID(s, m, commandStrings)

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid kick index.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}
	if index > len(misc.MemberInfoMap[userID].Warnings) || index < 0 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid kick index.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Fixes index for future use if it's 0
	if index != 0 {

		index = index - 1
	}

	//Pulls info on user
	userMem, err := s.User(userID)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Checks if user is in memberInfo, giving error if not
	if misc.MemberInfoMap[userID] == nil {

		// Prints error
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in memberInfo.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Prints success
	_, err = s.ChannelMessageSend(m.ChannelID, "Success. Removed kick `" + misc.MemberInfoMap[userID].Kicks[index] +
		"` from " + userMem.Username + "#" + userMem.Discriminator)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Removes warning from map
	misc.MapMutex.Lock()
	misc.MemberInfoMap[userID].Kicks = append(misc.MemberInfoMap[userID].Kicks[:index], misc.MemberInfoMap[userID].Kicks[index+1:]...)
	misc.MapMutex.Unlock()

	// Writes new map to storage
	misc.MemberInfoWrite(misc.MemberInfoMap)
}

func removeBanCommand(s *discordgo.Session, m *discordgo.Message) {

	// Puts entire message in lowercase
	messageLowercase := strings.ToLower(m.Content)

	//Separates every word in the message and puts it in a slice
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and index. Else prints error message
	if len(commandStrings) != 3 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Wrong amount of parameters.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID := misc.GetUserID(s, m, commandStrings)

	// Index checks
	index, err := strconv.Atoi(commandStrings[2])
	if err != nil {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid ban index.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}
	if index > len(misc.MemberInfoMap[userID].Warnings) || index < 0 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid ban index.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Fixes index for future use if it's 0
	if index != 0 {

		index = index - 1
	}

	//Pulls info on user
	userMem, err := s.User(userID)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Checks if user is in memberInfo, giving error if not
	if misc.MemberInfoMap[userID] == nil {

		// Prints error
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User does not exist in memberInfo.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Prints success
	_, err = s.ChannelMessageSend(m.ChannelID, "Success. Removed ban `" + misc.MemberInfoMap[userID].Bans[index] +
		"` from " + userMem.Username + "#" + userMem.Discriminator)
	if err != nil {
		fmt.Println("Error:", err)
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
//	})
//	add(&command{
//		execute:  removeKickCommand,
//		trigger:  "removekick",
//		desc:     "Removes a user kick whois text",
//		elevated: true,
//	})
//	add(&command{
//		execute:  removeBanCommand,
//		trigger:  "removeban",
//		desc:     "Removes a user ban whois text",
//		elevated: true,
//	})
//}