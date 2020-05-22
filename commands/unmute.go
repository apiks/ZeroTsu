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

// Unmutes a user and updates their memberInfo entry
func unmuteCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		mutedRoleID string

		muteFlag bool
		tookRole bool
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	if guildSettings.GetMutedRole() == (entities.Role{}) || guildSettings.GetMutedRole().GetID() == "" {
		// Pulls info on server roles
		deb, err := s.GuildRoles(m.GuildID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}

		// Checks by string for a muted role
		for _, role := range deb {
			if strings.ToLower(role.Name) == "muted" || strings.ToLower(role.Name) == "t-mute" {
				mutedRoleID = role.ID
				break
			}
		}
	} else {
		mutedRoleID = guildSettings.GetMutedRole().GetID()
	}

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	if len(commandStrings) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"unmute [@user, userID, or username#discrim]` format.\n\n"+
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

	// Pulls info on user
	userMem, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		userMem, err = s.GuildMember(m.GuildID, userID)
		if err != nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: Username is not in the server. Cannot unmute.")
			if err != nil {
				return
			}
			return
		}
	}

	// Goes through every muted user from PunishedUsers and if the user is in it, confirms that user is muted
	punishedUsers := db.GetGuildPunishedUsers(m.GuildID)
	if len(punishedUsers) == 0 {
		_, err = s.ChannelMessageSend(m.ChannelID, "No mutes found.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Removes the mute if it finds it
	for _, user := range punishedUsers {
		if user.GetID() == userID {
			muteFlag = true

			if user.GetUnbanDate() == common.NilTime {
				err = db.SetGuildPunishedUser(m.GuildID, entities.NewPunishedUsers("", "", time.Time{}, time.Time{}))
				if err != nil {
					common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
					return
				}
			} else {
				err = db.SetGuildPunishedUser(m.GuildID, entities.NewPunishedUsers(user.GetID(), user.GetUsername(), user.GetUnbanDate(), time.Time{}))
				if err != nil {
					common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
					return
				}
			}

			break
		}
	}

	// Check if the user is muted using other means
	if !muteFlag {
		if mutedRoleID != "" {
			for _, userRoleID := range userMem.Roles {
				if userRoleID == mutedRoleID {
					muteFlag = true
					break
				}
			}
		} else {
			// Pulls info on server roles
			deb, err := s.GuildRoles(m.GuildID)
			if err != nil {
				common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}

			// Checks by string for a muted role
			for _, role := range deb {
				if strings.ToLower(role.Name) == "muted" || strings.ToLower(role.Name) == "t-mute" {
					for _, userRoleID := range userMem.Roles {
						if userRoleID == mutedRoleID {
							muteFlag = true
							break
						}
					}
					break
				}
			}
		}
	}

	if !muteFlag {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("__%s#%s__ is not muted.", userMem.User.Username, userMem.User.Discriminator))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Removes the mute
	if mutedRoleID != "" {
		_ = s.GuildMemberRoleRemove(m.GuildID, userID, mutedRoleID)
		tookRole = true
	} else {
		// Pulls info on server roles
		deb, err := s.GuildRoles(m.GuildID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}

		// Checks by string for a muted role
		for _, role := range deb {
			if strings.ToLower(role.Name) == "muted" || strings.ToLower(role.Name) == "t-mute" {
				_ = s.GuildMemberRoleRemove(m.GuildID, userID, role.ID)
				tookRole = true
				break
			}
		}
	}

	if !tookRole {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: This server does not have a set muted role. Please use `%vsetmuted [Role ID]` before trying this command again.", guildSettings.GetPrefix()))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Saves time of unmute command usage
	t := time.Now()

	// Checks if user is in memberInfo and fetches them
	mem := db.GetGuildMember(m.GuildID, userID)
	if mem.GetID() == "" {
		var user *discordgo.User

		if userMem != nil {
			user = userMem.User
		} else {
			user, err = s.User(userID)
			if err != nil {
				_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in the server, internal database and cannot fetch manually either. Cannot unmute until user joins the server.")
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

	// Set member unmute date
	mem.SetUnmuteDate(t.Format("2006-01-02 15:04:05"))

	// Write
	db.SetGuildMember(m.GuildID, mem)

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("__%s#%s__ has been unmuted.", userMem.User.Username, userMem.User.Discriminator))
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Sends an embed message to bot-log if possible
	if guildSettings.BotLog != (entities.Cha{}) {
		if guildSettings.BotLog.GetID() != "" {
			_ = embeds.AutoPunishmentRemoval(s, mem, guildSettings.BotLog.GetID(), "unmuted", m.Author)
		}
	}
}

func init() {
	Add(&Command{
		Execute:    unmuteCommand,
		Trigger:    "unmute",
		Aliases:    []string{"unshut"},
		Desc:       "Unmutes a user",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
