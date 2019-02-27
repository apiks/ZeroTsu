package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Verifies a user with a reddit username and gives them the verified role
func verifyCommand(s *discordgo.Session, m *discordgo.Message) {

	var roleID string

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and reddit username.)
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"verify [@user, userID, or username#discrim] [redditUsername]`\n\n" +
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
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
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: User is not in the server. Cannot verify user until they rejoin the server.")
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

	// Add reddit username in map
	misc.MapMutex.Lock()
	if misc.MemberInfoMap[userID] != nil {

		// Stores time of verification
		t := time.Now()
		z, _ := t.Zone()
		ver := t.Format("2006-01-02 15:04:05") + " " + z

		// Sets verification variables
		misc.MemberInfoMap[userID].RedditUsername = redditUsername
		misc.MemberInfoMap[userID].VerifiedDate = ver
	} else {

		// Initializes user in memberInfo.json
		misc.InitializeUser(userMem)

		// Stores time of verification
		t := time.Now()
		z, _ := t.Zone()
		ver := t.Format("2006-01-02 15:04:05") + " " + z

		// Sets verification variables
		misc.MemberInfoMap[userID].RedditUsername = redditUsername
		misc.MemberInfoMap[userID].VerifiedDate = ver
	}
	misc.MapMutex.Unlock()

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

	err = verifyEmbed(s, m, userMem, redditUsername)
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

func verifyEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.Member, username string) error {

	var embedMess      discordgo.MessageEmbed

	// Sets punishment embed color
	embedMess.Color = 0x00ff00
	embedMess.Title = fmt.Sprintf("Successfuly verified %v#%v with /u/%v", mem.User.Username, mem.User.Discriminator, username)

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	return err
}

// Unverifies a user
func unverifyCommand(s *discordgo.Session, m *discordgo.Message) {

	var roleID string

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and reddit username.)
	if len(commandStrings) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"unverify [@user, userID, or username#discrim]`\n\n" +
			"Note: If using username#discrim you can have spaces in the username.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
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

	// Remove reddit username from map
	misc.MapMutex.Lock()
	if misc.MemberInfoMap[userID] != nil {

		// Sets verification variables
		misc.MemberInfoMap[userID].RedditUsername = ""
		misc.MemberInfoMap[userID].VerifiedDate = ""
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User is not in memberInfo. Cannot unverify user until they join the server.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
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
	err = s.GuildMemberRoleRemove(config.ServerID, userID, roleID)
	if err != nil {
		return
	}

	err = unverifyEmbed(s, m, commandStrings[1])
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

func unverifyEmbed(s *discordgo.Session, m *discordgo.Message, mem string) error {

	var embedMess      discordgo.MessageEmbed

	// Sets punishment embed color
	embedMess.Color = 0x00ff00
	embedMess.Title = fmt.Sprintf("Successfuly unverified %v", mem)

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	return err
}

func init() {
	add(&command{
		execute:  verifyCommand,
		trigger:  "verify",
		desc:     "Verifies a user with a reddit username.",
		elevated: true,
		category: "misc",
	})
	add(&command{
		execute:  unverifyCommand,
		trigger:  "unverify",
		desc:     "Unverifies a user.",
		elevated: true,
		category: "misc",
	})
}