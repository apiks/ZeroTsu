package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/functionality"
)

// Verifies a user with a reddit username and gives them the verified role
func verifyCommand(s *discordgo.Session, m *discordgo.Message) {
	if config.Website == "" {
		return
	}

	var roleID string

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Checks if there's enough parameters (command, user and reddit username.)
	if len(commandStrings) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"verify [@user, userID, or username#discrim] [redditUsername]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := common.GetUserID(m, commandStrings)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	// Checks if user is in memberInfo and fetches them
	mem := db.GetGuildMember(m.GuildID, userID)
	if mem.GetID() == "" {
		var user *discordgo.User

		if userMem != nil {
			user = userMem.User
		} else {
			user, err = s.User(userID)
			if err != nil {
				_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in the server, internal database and cannot fetch manually either. Cannot verify until user joins the server.")
				if err != nil {
					common.LogError(s, guildSettings.BotLog, err)
					return
				}
				return
			}
		}

		// Initializes user if he doesn't exist in memberInfo but is in server
		functionality.InitializeUser(user, m.GuildID)

		mem = db.GetGuildMember(m.GuildID, userID)
		if mem.GetID() == "" {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, fmt.Errorf("error: member object is empty"))
			return
		}
	}

	// Stores time of verification and adds reddit username
	t := time.Now()
	z, _ := t.Zone()
	ver := t.Format("2006-01-02 15:04:05") + " " + z
	mem = mem.SetRedditUsername(redditUsername)
	mem = mem.SetVerifiedDate(ver)

	// Write
	db.SetGuildMember(m.GuildID, mem)

	// Puts all server roles in roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Stores time of verification and adds to verification stats
	db.AddGuildVerifiedStat(m.GuildID, t.Format(common.ShortDateFormat), 1)

	// Sends warning embed message to channel
	if userMem == nil {
		userMem = &discordgo.Member{GuildID: m.GuildID, User: &discordgo.User{ID: mem.GetID(), Username: mem.GetUsername(), Discriminator: mem.GetDiscrim()}}
	}

	err = embeds.Verification(s, m, userMem, redditUsername)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Unverifies a user
func unverifyCommand(s *discordgo.Session, m *discordgo.Message) {
	if config.Website == "" {
		return
	}

	var roleID string

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	// Checks if there's enough parameters (command, user and reddit username.)
	if len(commandStrings) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"unverify [@user, userID, or username#discrim]`\n\n"+
			"Note: If using username#discrim you can have spaces in the username.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls userID from 2nd parameter of commandStrings, else print error
	userID, err := common.GetUserID(m, commandStrings)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Pulls info on user
	userMem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		userMem, _ = s.GuildMember(m.GuildID, userID)
	}

	// Checks if user is in memberInfo and fetches them
	mem := db.GetGuildMember(m.GuildID, userID)
	if mem.GetID() == "" {
		if userMem == nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in the server and internal database. Cannot unverify until user joins the server.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}

		// Initializes user if he doesn't exist in memberInfo but is in server
		functionality.InitializeUser(userMem.User, m.GuildID)

		mem = db.GetGuildMember(m.GuildID, userID)
		if mem.GetID() == "" {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, fmt.Errorf("error: member object is empty"))
			return
		}
	}

	// Remove reddit username from member
	mem = mem.SetRedditUsername("")
	mem = mem.SetVerifiedDate("")

	// Write
	db.SetGuildMember(m.GuildID, mem)

	// Puts all server roles in roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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

	// Removes from verification stats
	t := time.Now()
	db.AddGuildVerifiedStat(m.GuildID, t.Format(common.ShortDateFormat), -1)

	err = embeds.Verification(s, m, userMem)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func init() {
	Add(&Command{
		Execute:    verifyCommand,
		Trigger:    "verify",
		Desc:       "Verifies a user with a reddit username",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
	Add(&Command{
		Execute:    unverifyCommand,
		Trigger:    "unverify",
		Desc:       "Unverifies a user",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
