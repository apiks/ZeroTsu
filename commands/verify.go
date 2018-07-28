package commands

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
	"github.com/r-anime/ZeroTsu/config"
)

// Verifies a user with a reddit username and gives them the verified role
func verifyCommand(s *discordgo.Session, m *discordgo.Message) {

	var roleID string

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and reddit username.)
	if len(commandStrings) != 3 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+"verify [@user or userID] [redditUsername]`")
		if err != nil {

			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
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

			_, err := s.ChannelMessageSend(m.ChannelID, "Error: User is not in the server. Cannot verify user.")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
				if err != nil {

					return
				}
				return
			}
			return
		}
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
	}

	// Writes modified memberInfo map to storage
	misc.MemberInfoWrite(misc.MemberInfoMap)

	// Puts all server roles in roles
	roles, err := s.GuildRoles(config.ServerID)
	if err != nil {

		misc.CommandErrorHandler(s, m, err)
		return
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

		misc.CommandErrorHandler(s, m, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success. Verified "+userMem.User.Username+"#"+userMem.User.Discriminator+" with "+redditUsername)
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
//		execute:  verifyCommand,
//		trigger:  "verify",
//		desc:     "Verifies a user with a reddit username.",
//		elevated: true,
//	})
//}