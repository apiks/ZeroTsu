package commands

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/functionality"
)

// Verifies a user with a reddit username and gives them the verified role
func verifyCommand(s *discordgo.Session, m *discordgo.Message) {

	var roleID string

	if config.Website == "" {
		return
	}

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Checks if there's enough parameters (command, user and reddit username.)
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"verify [@user, userID, or username#discrim] [redditUsername]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := functionality.GetUserID(m, commandStrings)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
	functionality.Mutex.Lock()
	if memberInfoUser, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]; ok {

		// Stores time of verification
		t := time.Now()
		z, _ := t.Zone()
		ver := t.Format("2006-01-02 15:04:05") + " " + z

		// Sets verification variables
		memberInfoUser.RedditUsername = redditUsername
		memberInfoUser.VerifiedDate = ver
	} else if userMem != nil {

		// Initializes user in memberInfo.json
		functionality.InitializeMember(userMem, m.GuildID)

		// Stores time of verification
		t := time.Now()
		z, _ := t.Zone()
		ver := t.Format("2006-01-02 15:04:05") + " " + z

		// Sets verification variables
		memberInfoUser.RedditUsername = redditUsername
		memberInfoUser.VerifiedDate = ver
	} else {
		functionality.Mutex.Unlock()
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User is not in the server _and_ internal database. Cannot verify user until they rejoin the server.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Writes modified memberInfo map to storage
	_ = functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	functionality.Mutex.Unlock()

	// Puts all server roles in roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Stores time of verification
	t := time.Now()
	// Adds to verified stats
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].VerifiedStats[t.Format(functionality.DateFormat)]++
	functionality.Mutex.Unlock()

	if userMem == nil {
		return
	}

	err = functionality.VerifyEmbed(s, m, userMem, redditUsername)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Unverifies a user
func unverifyCommand(s *discordgo.Session, m *discordgo.Message) {

	if config.Website == "" {
		return
	}

	var roleID string

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Checks if there's enough parameters (command, user and reddit username.)
	if len(commandStrings) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"unverify [@user, userID, or username#discrim]`\n\n"+
			"Note: If using username#discrim you can have spaces in the username.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := functionality.GetUserID(m, commandStrings)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Remove reddit username from map
	functionality.Mutex.Lock()
	if memberInfoUser, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]; ok {
		// Sets verification variables
		memberInfoUser.RedditUsername = ""
		memberInfoUser.VerifiedDate = ""
	} else {
		functionality.Mutex.Unlock()
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User is not in internal database. Cannot unverify user until they join the server.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Writes modified memberInfo map to storage
	_ = functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	functionality.Mutex.Unlock()

	// Puts all server roles in roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
	functionality.Mutex.Lock()
	functionality.GuildMap[m.GuildID].VerifiedStats[t.Format(functionality.DateFormat)]--
	functionality.Mutex.Unlock()

	err = functionality.UnverifyEmbed(s, m, commandStrings[1])
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    verifyCommand,
		Trigger:    "verify",
		Desc:       "Verifies a user with a reddit username",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
	functionality.Add(&functionality.Command{
		Execute:    unverifyCommand,
		Trigger:    "unverify",
		Desc:       "Unverifies a user",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
