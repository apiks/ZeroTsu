package commands

import (
	"strings"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
	"github.com/r-anime/ZeroTsu/config"
)

// Verifies a user with a reddit username and gives them the verified role
func verifyCommand(s *discordgo.Session, m *discordgo.Message) {

	// Puts entire message in lowercase
	messageLowercase := strings.ToLower(m.Content)

	// Separates every word in the messageLowercase and puts it in a slice
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and reddit username.)
	if len(commandStrings) != 3 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + "verify [@user or userID] [redditUsername]`")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID := misc.GetUserID(s, m, commandStrings)

	// Pulls the reddit username from the third parameter
	redditUsername := commandStrings[2]

	// Trims the reddit username if it's done with /u/ or u/
	if strings.HasPrefix(redditUsername, "/u/") {

		redditUsername = strings.TrimPrefix(redditUsername, "/u/")
	} else if strings.HasPrefix(redditUsername, "u/") {

		redditUsername = strings.TrimPrefix(redditUsername, "u/")
	}

	// Pulls info on user
	userMem, err := s.State.Member(config.ServerID, userID)
	if err != nil {
		userMem, err = s.GuildMember(config.ServerID, userID)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	if userMem == nil {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User is not in the server.")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Add reddit username in map
	if misc.MemberInfoMap[userID] != nil {

		// Stores time of verification
		t := time.Now()
		z, _ := t.Zone()
		ver := t.Format("2006-01-02 15:04:05") + " " + z

		// Sets verification variables
		misc.MapMutex.Lock()
		misc.MemberInfoMap[userID].RedditUsername = redditUsername
		misc.MemberInfoMap[userID].VerifiedDate = ver
		misc.MapMutex.Unlock()

		// Writes modified memberInfo map to storage
		misc.MemberInfoWrite(misc.MemberInfoMap)

		_, err := s.ChannelMessageSend(m.ChannelID, "Success. Verified "+userMem.User.Username + "#" + userMem.User.Discriminator +" with "+redditUsername)
		if err != nil {
			fmt.Println("Error:", err)
		}
	} else {

		// Initializes user in memberInfo.json
		misc.InitializeUser(userMem)

		// Stores time of verification
		t := time.Now()
		z, _ := t.Zone()
		ver := t.Format("2006-01-02 15:04:05") + " " + z

		// Sets verification variables
		misc.MapMutex.Lock()
		misc.MemberInfoMap[userID].RedditUsername = redditUsername
		misc.MemberInfoMap[userID].VerifiedDate = ver
		misc.MapMutex.Unlock()

		// Writes modified memberInfo map to storage
		misc.MemberInfoWrite(misc.MemberInfoMap)

		_, err = s.ChannelMessageSend(m.ChannelID, "Success. Verified "+ userMem.User.Mention() + " with " + redditUsername)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}

	var roleID string

	// Puts all server roles in roles
	roles, err := s.GuildRoles(config.ServerID)
	if err != nil {

		fmt.Println("Error:", err)
	}

	// Fetches ID of Verified role and finds the correct one
	for i := 0; i < len(roles); i++ {
		if roles[i].Name == "Verified" {

			roleID = roles[i].ID
		}
	}

	// Assigns verified role to user
	err = s.GuildMemberRoleAdd(config.ServerID, userID, roleID)
	if err != nil {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Verified role not found.")
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}

//func init() {
//	add(&command{
//		execute:  verifyCommand,
//		trigger:  "verify",
//		desc:     "Verifies a user with a reddit username.",
//		elevated: true,
//	})
//}