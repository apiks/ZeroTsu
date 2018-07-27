package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
	"github.com/r-anime/ZeroTsu/config"
)

func avatarCommand(s *discordgo.Session, m *discordgo.Message) {

	// Puts entire message in lowercase
	messageLowercase := strings.ToLower(m.Content)

	// Separates every word in the message and puts it in a slice
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and index.) Else prints error message
	if len(commandStrings) == 1 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Incorrect usage: Try `" + config.BotPrefix + "avatar [@user or userID]`")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	} else if len(commandStrings) != 2 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Wrong amount of parameters.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID := misc.GetUserID(s, m, commandStrings)

	// Fetches user from server
	mem, err := s.User(userID)
	if err != nil {
		return
	}

	// Saves the avatar URL to avatar variable with image size 256
	avatar := mem.AvatarURL("256")

	_, err = s.ChannelMessageSend(m.ChannelID, avatar)
	if err != nil {

		fmt.Println("Error:", err)
	}
}

func init() {
	add(&command{
		execute: avatarCommand,
		trigger: "avatar",
		desc:    "Show user avatar.",
	})
}