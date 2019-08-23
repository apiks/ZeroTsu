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

	if config.Website == "" {
		return
	}

	misc.MapMutex.Lock()
	misc.LoadDB(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and reddit username.)
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"verify [@user, userID, or username#discrim] [redditUsername]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := misc.GetUserID(m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Pulls info on user
	userMem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		userMem, _ = s.GuildMember(m.GuildID, userID)
	}

	// Pulls the reddit username from the third parameter
	redditUsername := commandStrings[2]

	// Trims the reddit username if it's done with /u/ or u/
	if strings.HasPrefix(redditUsername, "/u/") {
		redditUsername = strings.TrimPrefix(redditUsername, "/u/")
	} else if strings.HasPrefix(redditUsername, "u/") {
		redditUsername = strings.TrimPrefix(redditUsername, "u/")
	}

	// Add reddit username in map
	misc.MapMutex.Lock()
	misc.LoadDB(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	if _, ok := misc.GuildMap[m.GuildID].MemberInfoMap[userID]; ok {

		// Stores time of verification
		t := time.Now()
		z, _ := t.Zone()
		ver := t.Format("2006-01-02 15:04:05") + " " + z

		// Sets verification variables
		misc.GuildMap[m.GuildID].MemberInfoMap[userID].RedditUsername = redditUsername
		misc.GuildMap[m.GuildID].MemberInfoMap[userID].VerifiedDate = ver
	} else if userMem != nil {

		// Initializes user in memberInfo.json
		misc.InitializeUser(userMem, m.GuildID)

		// Stores time of verification
		t := time.Now()
		z, _ := t.Zone()
		ver := t.Format("2006-01-02 15:04:05") + " " + z

		// Sets verification variables
		misc.GuildMap[m.GuildID].MemberInfoMap[userID].RedditUsername = redditUsername
		misc.GuildMap[m.GuildID].MemberInfoMap[userID].VerifiedDate = ver
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User is not in the server _and_ MemberInfo. Cannot verify user until they rejoin the server.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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

	// Writes modified memberInfo map to storage
	misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	misc.MapMutex.Unlock()

	// Puts all server roles in roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Fetches ID of Verified role and finds the correct one
	for i := 0; i < len(roles); i++ {
		if roles[i].Name == "Verified" {
			roleID = roles[i].ID
		}
	}

	// Assigns verified role to user
	err = s.GuildMemberRoleAdd(m.GuildID, userID, roleID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Stores time of verification
	t := time.Now()
	// Adds to verified stats
	misc.MapMutex.Lock()
	misc.LoadDB(misc.GuildMap[m.GuildID].VerifiedStats, m.GuildID)
	misc.GuildMap[m.GuildID].VerifiedStats[t.Format(misc.DateFormat)]++
	misc.MapMutex.Unlock()

	if userMem == nil {
		return
	}

	err = verifyEmbed(s, m, userMem, redditUsername)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func verifyEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.Member, username string) error {

	var embedMess discordgo.MessageEmbed

	// Sets punishment embed color
	embedMess.Color = 0x00ff00
	embedMess.Title = fmt.Sprintf("Successfully verified %v#%v with /u/%v", mem.User.Username, mem.User.Discriminator, username)

	// Sends embed in channel
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embedMess)
	return err
}

// Unverifies a user
func unverifyCommand(s *discordgo.Session, m *discordgo.Message) {

	if config.Website == "" {
		return
	}

	var roleID string

	misc.MapMutex.Lock()
	misc.LoadDB(misc.GuildMap[m.GuildID].GuildConfig, m.GuildID)
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if there's enough parameters (command, user and reddit username.)
	if len(commandStrings) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"unverify [@user, userID, or username#discrim]`\n\n"+
			"Note: If using username#discrim you can have spaces in the username.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := misc.GetUserID(m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Remove reddit username from map
	misc.MapMutex.Lock()
	misc.LoadDB(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	if _, ok := misc.GuildMap[m.GuildID].MemberInfoMap[userID]; ok {
		// Sets verification variables
		misc.GuildMap[m.GuildID].MemberInfoMap[userID].RedditUsername = ""
		misc.GuildMap[m.GuildID].MemberInfoMap[userID].VerifiedDate = ""
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User is not in memberInfo. Cannot unverify user until they join the server.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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

	// Writes modified memberInfo map to storage
	misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	misc.MapMutex.Unlock()

	// Puts all server roles in roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Fetches ID of Verified role and finds the correct one
	for i := 0; i < len(roles); i++ {
		if roles[i].Name == "Verified" {
			roleID = roles[i].ID
		}
	}

	// Removes verified role from user
	err = s.GuildMemberRoleRemove(m.GuildID, userID, roleID)
	if err != nil {
		return
	}

	// Stores time of verification
	t := time.Now()
	// Removes from verified stats
	misc.MapMutex.Lock()
	misc.LoadDB(misc.GuildMap[m.GuildID].VerifiedStats, m.GuildID)
	misc.GuildMap[m.GuildID].VerifiedStats[t.Format(misc.DateFormat)]--
	misc.MapMutex.Unlock()

	err = unverifyEmbed(s, m, commandStrings[1])
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func unverifyEmbed(s *discordgo.Session, m *discordgo.Message, mem string) error {

	var embedMess discordgo.MessageEmbed

	// Sets punishment embed color
	embedMess.Color = 0x00ff00
	embedMess.Title = fmt.Sprintf("Successfully unverified %v", mem)

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
