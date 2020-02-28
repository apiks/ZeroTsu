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

// Unbans a user and updates their memberInfo entry
func unbanCommand(s *discordgo.Session, m *discordgo.Message) {

	var banFlag = false

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	if len(commandStrings) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"unban [@user, userID, or username#discrim]` format.\n\n"+
			"Note: this command supports username#discrim where username contains spaces.")
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

	user, err := s.User(userID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Goes through every banned user from PunishedUsers and if the user is in it, confirms that user is a temp ban
	punishedUsers := db.GetGuildPunishedUsers(m.GuildID)
	if len(punishedUsers) == 0 {
		_, err = s.ChannelMessageSend(m.ChannelID, "No bans found.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Removes the ban if it finds it
	for _, user := range punishedUsers {
		if user.GetID() == userID {
			banFlag = true

			if user.GetUnmuteDate() == (time.Time{}) {
				err = db.SetGuildPunishedUser(m.GuildID, entities.NewPunishedUsers("", "", time.Time{}, time.Time{}))
				if err != nil {
					common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
					return
				}
			} else {
				err = db.SetGuildPunishedUser(m.GuildID, entities.NewPunishedUsers(user.GetID(), user.GetUsername(), time.Time{}, user.GetUnmuteDate()))
				if err != nil {
					common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
					return
				}
			}

			break
		}
	}

	// Check if the user is banned using other means
	if !banFlag {
		bans, err := s.GuildBans(m.GuildID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		for _, ban := range bans {
			if ban.User.ID == userID {
				banFlag = true
				break
			}
		}
	}

	if !banFlag {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("__%s#%s__ is not banned.", user.Username, user.Discriminator))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Removes the ban
	err = s.GuildBanDelete(m.GuildID, userID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Saves time of unban command usage
	t := time.Now()

	// Checks if user is in memberInfo and fetches them
	mem := db.GetGuildMember(m.GuildID, userID)
	if mem.GetID() == "" {
		// Initializes user if he doesn't exist in memberInfo but is in server
		functionality.InitializeUser(user, m.GuildID)

		mem = db.GetGuildMember(m.GuildID, userID)
		if mem.GetID() == "" {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, fmt.Errorf("error: member object is empty"))
			return
		}
	}

	// Set member unban date
	mem = mem.SetUnbanDate(t.Format("2006-01-02 15:04:05"))

	// Write
	db.SetGuildMember(m.GuildID, mem)

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("__%s#%s__ has been unbanned.", user.Username, user.Discriminator))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Sends an embed message to bot-log if possible
	if guildSettings.BotLog != (entities.Cha{}) {
		if guildSettings.BotLog.GetID() != "" {
			_ = embeds.AutoPunishmentRemoval(s, mem, guildSettings.BotLog.GetID(), "unbanned", m.Author)
		}
	}
}

func init() {
	Add(&Command{
		Execute:    unbanCommand,
		Trigger:    "unban",
		Desc:       "Unbans a user",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
