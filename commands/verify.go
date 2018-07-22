package commands

import (
	"strings"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
	"github.com/r-anime/ZeroTsu/config"
)

func verifyCommand(s *discordgo.Session, m *discordgo.Message) {

	// Puts entire message in lowercase
	messageLowercase := strings.ToLower(m.Content)

	//Separates every word in the message and puts it in a slice
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and reddit username. Else prints error message
	if len(commandStrings) != 3 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Wrong amount of parameters.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Pulls the userID from the second parameter
	userID := commandStrings[1]

	// Trims fluff if it was a mention. Otherwise check if it's a correct user ID
	if strings.Contains(commandStrings[1], "<@") {

		userID = strings.TrimPrefix(userID, "<@")
		userID = strings.TrimSuffix(userID, ">")
	} else {

		_, err := strconv.ParseInt(userID, 10, 64)
		if len(userID) != 18 || err != nil {

			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid user ID.")
			if err != nil {
				fmt.Println("Error:", err)
			}
			return
		}
	}

	//Pulls the reddit username from the third parameter
	redditUsername := commandStrings[2]

	// Trims the reddit username if it's done with /u/ or u/
	if strings.HasPrefix(redditUsername, "/u/") {

		redditUsername = strings.TrimPrefix(redditUsername, "/u/")
	} else if strings.HasPrefix(redditUsername, "u/") {

		redditUsername = strings.TrimPrefix(redditUsername, "u/")
	}

	// Add reddit username in map
	misc.MapMutex.Lock()
	misc.MemberInfoMap[userID].RedditUsername = redditUsername
	misc.MapMutex.Unlock()

	// Writes modified memberInfo map to storage
	misc.MemberInfoWrite(misc.MemberInfoMap)

	// Initializes var roleID which will keep the Verified role ID
	var roleID string

	// Puts all server roles in roles variable
	roles, err := s.GuildRoles(config.ServerID)
	if err != nil {

		fmt.Println("Error:", err)

		// Fetches ID of Verified role and finds the correct one
		for i := 0; i < len(roles); i++ {
			if roles[i].Name == "Verified" {

				roleID = roles[i].ID
			}
		}

		// Assigns verified role to user
		err = s.GuildMemberRoleAdd(config.ServerID, userID, roleID)
		if err != nil {

			fmt.Println("Error:", err)
		}
	}
}

//func init() {
//	add(&command{
//		execute:  verifyCommand,
//		trigger:  "verify",
//		desc:     "Verifies a user with a reddit username",
//		elevated: true,
//	})
//}