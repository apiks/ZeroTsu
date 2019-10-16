package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Unmutes a user and updates their memberInfo entry
func unmuteCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		muteFlag bool
		guildMutedRoleID string
		tookRole bool
	)

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()
	if guildSettings.MutedRole != nil {
		if guildSettings.MutedRole.ID != "" {
			guildMutedRoleID = guildSettings.MutedRole.ID
		}
	}

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	if len(commandStrings) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"unmute [@user, userID, or username#discrim]` format.\n\n"+
			"Note: this command supports username#discrim where username contains spaces.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	userID, err := functionality.GetUserID(m, commandStrings)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Pulls info on user
	user, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		user, err = s.GuildMember(m.GuildID, userID)
		if err != nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "Error: User is not in the server. Cannot unmute.")
			if err != nil {
				return
			}
			return
		}
	}

	// Goes through every muted user from punishedUsers and if the user is in it, confirms that user is a mute
	functionality.MapMutex.Lock()
	if functionality.GuildMap[m.GuildID].PunishedUsers == nil || len(functionality.GuildMap[m.GuildID].PunishedUsers) == 0 {
		_, err = s.ChannelMessageSend(m.ChannelID, "No mutes found.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}

	for i := 0; i < len(functionality.GuildMap[m.GuildID].PunishedUsers); i++ {
		if functionality.GuildMap[m.GuildID].PunishedUsers[i].ID == userID {
			muteFlag = true
			zeroTimeValue := time.Time{}

			// Removes the mute from punishedUsers
			if functionality.GuildMap[m.GuildID].PunishedUsers[i].UnbanDate != zeroTimeValue {
				temp := functionality.PunishedUsers{
					ID:        functionality.GuildMap[m.GuildID].PunishedUsers[i].ID,
					User:      functionality.GuildMap[m.GuildID].PunishedUsers[i].User,
					UnbanDate: functionality.GuildMap[m.GuildID].PunishedUsers[i].UnbanDate,
				}
				functionality.GuildMap[m.GuildID].PunishedUsers[i] = &temp
			} else {
				functionality.GuildMap[m.GuildID].PunishedUsers = append(functionality.GuildMap[m.GuildID].PunishedUsers[:i], functionality.GuildMap[m.GuildID].PunishedUsers[i+1:]...)
			}
			break
		}
	}
	functionality.MapMutex.Unlock()

	// Check if the user is muted using other means
	if !muteFlag {
		if guildMutedRoleID != "" {
			for _, userRoleID := range user.Roles {
				if userRoleID == guildMutedRoleID {
					muteFlag = true
					break
				}
			}
		} else {
			// Pulls info on server roles
			deb, err := s.GuildRoles(m.GuildID)
			if err != nil {
				functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}

			// Checks by string for a muted role
			for _, role := range deb {
				if strings.ToLower(role.Name) == "muted" || strings.ToLower(role.Name) == "t-mute" {
					for _, userRoleID := range user.Roles {
						if userRoleID == guildMutedRoleID {
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
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("__%s#%s__ is not muted.", user.User.Username, user.User.Discriminator))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Removes the mute
	if guildMutedRoleID != "" {
		_ = s.GuildMemberRoleRemove(m.GuildID, userID, guildMutedRoleID)
		tookRole = true
	} else {
		// Pulls info on server roles
		deb, err := s.GuildRoles(m.GuildID)
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: This server does not have a set muted role. Please use `%vsetmuted [Role ID]` before trying this command again.", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Saves time of unmute command usage
	t := time.Now()

	// Updates unmute date in memberInfo.json entry if possible and writes to storage
	functionality.MapMutex.Lock()
	if _, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]; ok {
		functionality.GuildMap[m.GuildID].MemberInfoMap[userID].UnmuteDate = t.Format("2006-01-02 15:04:05")
		_ = functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	}
	_ = functionality.PunishedUsersWrite(functionality.GuildMap[m.GuildID].PunishedUsers, m.GuildID)
	functionality.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("__%v#%v__ has been unmuted.", user.User.Username, user.User.Discriminator))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Sends an embed message to bot-log if possible
	functionality.MapMutex.Lock()
	if _, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]; ok {
		if guildSettings.BotLog != nil {
			if guildSettings.BotLog.ID != "" {
				_ = functionality.UnmuteEmbed(s, functionality.GuildMap[m.GuildID].MemberInfoMap[userID], m.Author.Username, guildSettings.BotLog.ID)
			}
		}
	}
	functionality.MapMutex.Unlock()
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    unmuteCommand,
		Trigger:    "unmute",
		Desc:       "Unmutes a user",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
