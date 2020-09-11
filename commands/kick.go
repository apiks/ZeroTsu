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

// Kicks a user from the server with a reason
func kickCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		userID        string
		reason        string
		kickTimestamp = entities.NewPunishment("", "", time.Time{})
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 3)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"kick [@user, userID, or username#discrim] [reason]*` format.\n\n* is optional"+
			"\n\nNote: If using username#discrim you cannot have spaces in the username. It must be a single word.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	userID, err := common.GetUserID(m, commandStrings)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	if len(commandStrings) == 3 {
		reason = commandStrings[2]
		// Checks if the reason contains a mention and finds the actual name instead of ID
		reason = common.MentionParser(s, reason, m.GuildID)
	} else {
		reason = "[No reason given]"
	}

	// Pulls info on user
	userMem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		userMem, err = s.GuildMember(m.GuildID, userID)
		if err != nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: Invalid user or user not found in server. Cannot kick user.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}

	// Checks if user has a privileged role
	if functionality.HasElevatedPermissions(s, userID, m.GuildID) {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Target user has a privileged role. Cannot kick.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if user is in memberInfo and handles them
	mem := db.GetGuildMember(m.GuildID, userID)
	if mem.GetID() == "" {
		if userMem == nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in server. Cannot kick until user joins the server.")
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

	// Adds kick reason to user memberInfo info
	mem = mem.AppendToKicks(reason)

	// Adds timestamp for that kick
	t, err := m.Timestamp.Parse()
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	kickTimestamp = kickTimestamp.SetTimestamp(t)
	kickTimestamp = kickTimestamp.SetPunishment(reason)
	kickTimestamp = kickTimestamp.SetPunishmentType("Kick")
	mem = mem.AppendToTimestamps(kickTimestamp)

	// Write
	db.SetGuildMember(m.GuildID, mem)

	// Fetches the guild for the Name
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		guild, err = s.Guild(m.GuildID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	// Sends message to user DMs if possible
	dm, _ := s.UserChannelCreate(userID)
	_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("You have been kicked from **%s**:\n**\"**%s**\"**", guild.Name, reason))

	// Kicks the user from the server with a reason
	err = s.GuildMemberDeleteWithReason(m.GuildID, userID, reason)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Sends embed channel message
	err = embeds.PunishmentAddition(s, m, userMem, "kick", "kicked", reason, m.ChannelID, nil)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Sends embed bot-log message
	if guildSettings.BotLog == (entities.Cha{}) {
		return
	}
	if guildSettings.BotLog.GetID() == "" {
		return
	}
	_ = embeds.PunishmentAddition(s, m, userMem, "kick", "kicked", reason, guildSettings.BotLog.GetID(), nil)
}

func init() {
	Add(&Command{
		Execute:    kickCommand,
		Trigger:    "kick",
		Aliases:    []string{"k", "yeet", "yut"},
		Desc:       "Kicks a user from the server and logs reason",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
