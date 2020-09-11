package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"github.com/r-anime/ZeroTsu/entities"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Adds a warning to a specific user in memberInfo.json without telling them
func addWarningCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		warning          string
		warningTimestamp = entities.NewPunishment("", "", time.Time{})
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	cmdStrs := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 3)

	if len(cmdStrs) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"addwarning [@user, userID, or username#discrim] [warning]`\n\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	userID, err := common.GetUserID(m, cmdStrs)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	warning = cmdStrs[2]
	// Checks if the warning contains a mention and finds the actual name instead of ID
	warning = common.MentionParser(s, warning, m.GuildID)

	// Pulls info on user
	userMem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		userMem, _ = s.GuildMember(m.GuildID, userID)
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
				_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in the server, internal database and cannot fetch manually either. Cannot warn until user joins the server.")
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

	// Appends warning to user in memberInfo
	mem = mem.AppendToWarnings(warning)

	// Adds timestamp for that warning
	t, err := m.Timestamp.Parse()
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	warningTimestamp = warningTimestamp.SetTimestamp(t)
	warningTimestamp = warningTimestamp.SetPunishment(warning)
	warningTimestamp = warningTimestamp.SetPunishmentType("Warning")
	mem = mem.AppendToTimestamps(warningTimestamp)

	// Write
	db.SetGuildMember(m.GuildID, mem)

	// Sends warning embed message to channel
	if userMem == nil {
		userMem = &discordgo.Member{GuildID: m.GuildID, User: &discordgo.User{ID: mem.GetID(), Username: mem.GetUsername(), Discriminator: mem.GetDiscrim()}}
	}

	err = embeds.PunishmentAddition(s, m, userMem, "warning", "warned", warning, m.ChannelID, nil, true)
	if err != nil {
		return
	}
}

// Issues a warning to a specific user in memberInfo.json wand tells them
func issueWarningCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		warning          string
		warningTimestamp = entities.NewPunishment("", "", time.Time{})
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	cmdStrs := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 3)

	if len(cmdStrs) != 3 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"issuewarning [@user, userID, or username#discrim] [warning]`\n"+
			"Note: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls the guild name early on purpose
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		guild, err = s.Guild(m.GuildID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	userID, err := common.GetUserID(m, cmdStrs)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	warning = cmdStrs[2]
	// Checks if the warning contains a mention and finds the actual name instead of ID
	warning = common.MentionParser(s, warning, m.GuildID)

	// Pulls info on user
	userMem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		userMem, _ = s.GuildMember(m.GuildID, userID)
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
				_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in the server, internal database and cannot fetch manually either. Cannot ban until user joins the server.")
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

	// Appends warning to user in memberInfo
	mem = mem.AppendToWarnings(warning)

	// Adds timestamp for that warning
	t, err := m.Timestamp.Parse()
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	warningTimestamp = warningTimestamp.SetTimestamp(t)
	warningTimestamp = warningTimestamp.SetPunishment(warning)
	warningTimestamp = warningTimestamp.SetPunishmentType("Warning")
	mem = mem.AppendToTimestamps(warningTimestamp)

	// Write
	db.SetGuildMember(m.GuildID, mem)

	// Sends message in DMs that they have been warned if able
	dm, _ := s.UserChannelCreate(userID)
	_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("You have been warned on **%s**:\n**\"**%s**\"**", guild.Name, warning))

	// Sends warning embed message to channel
	if userMem == nil {
		userMem = &discordgo.Member{GuildID: m.GuildID, User: &discordgo.User{ID: mem.GetID(), Username: mem.GetUsername(), Discriminator: mem.GetDiscrim()}}
	}

	err = embeds.PunishmentAddition(s, m, userMem, "warning", "warned", warning, m.ChannelID, nil)
	if err != nil {
		return
	}
}

func init() {
	Add(&Command{
		Execute:    addWarningCommand,
		Trigger:    "addwarning",
		Desc:       "Adds a warning to a user without messaging them",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
	Add(&Command{
		Execute:    issueWarningCommand,
		Trigger:    "issuewarning",
		Desc:       "Issues a warning to a user and messages them",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
