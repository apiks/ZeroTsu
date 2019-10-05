package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/misc"
	"strings"
	"time"
)

// Unmutes a user and updates their memberInfo entry
func unmuteCommand(s *discordgo.Session, m *discordgo.Message) {

	var muteFlag = false
	var guildMutedRoleID string

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	if misc.GuildMap[m.GuildID].GuildConfig.MutedRole != nil {
		guildMutedRoleID = misc.GuildMap[m.GuildID].GuildConfig.MutedRole.ID
	}
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"unmute [@user, userID, or username#discrim]` format.\n\n"+
			"Note: this command supports username#discrim where username contains spaces.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	userID, err := misc.GetUserID(m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Pulls info on user
	user, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		user, err = s.GuildMember(m.GuildID, userID)
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, "Error: User is not in the server. Cannot unmute.")
			if err != nil {
				return
			}
			return
		}
	}

	// Goes through every muted user from punishedUsers and if the user is in it, confirms that user is a mute
	misc.MapMutex.Lock()
	if len(misc.GuildMap[m.GuildID].PunishedUsers) == 0 {
		_, err = s.ChannelMessageSend(m.ChannelID, "No mutes found.")
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

	for i := 0; i < len(misc.GuildMap[m.GuildID].PunishedUsers); i++ {
		if misc.GuildMap[m.GuildID].PunishedUsers[i].ID == userID {
			muteFlag = true
			zeroTimeValue := time.Time{}

			// Removes the mute from punishedUsers
			if misc.GuildMap[m.GuildID].PunishedUsers[i].UnbanDate != zeroTimeValue {
				temp := misc.PunishedUsers {
					ID:         misc.GuildMap[m.GuildID].PunishedUsers[i].ID,
					User:       misc.GuildMap[m.GuildID].PunishedUsers[i].User,
					UnbanDate: misc.GuildMap[m.GuildID].PunishedUsers[i].UnbanDate,
				}
				misc.GuildMap[m.GuildID].PunishedUsers[i] = temp
			} else {
				misc.GuildMap[m.GuildID].PunishedUsers = append(misc.GuildMap[m.GuildID].PunishedUsers[:i], misc.GuildMap[m.GuildID].PunishedUsers[i+1:]...)
			}
			break
		}
	}
	misc.MapMutex.Unlock()

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
				misc.CommandErrorHandler(s, m, err, guildBotLog)
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
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("__%v#%v__ is not muted.", user.User.Username, user.User.Discriminator))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Removes the mute
	if guildMutedRoleID != "" {
		_ = s.GuildMemberRoleRemove(m.GuildID, userID, guildMutedRoleID)
	} else {
		// Pulls info on server roles
		deb, err := s.GuildRoles(m.GuildID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}

		// Checks by string for a muted role
		for _, role := range deb {
			if strings.ToLower(role.Name) == "muted" || strings.ToLower(role.Name) == "t-mute" {
				_ = s.GuildMemberRoleRemove(m.GuildID, userID, role.ID)
				break
			}
		}
	}

	// Saves time of unmute command usage
	t := time.Now()

	// Updates unmute date in memberInfo.json entry if possible and writes to storage
	misc.MapMutex.Lock()
	if _, ok := misc.GuildMap[m.GuildID].MemberInfoMap[userID]; ok {
		misc.GuildMap[m.GuildID].MemberInfoMap[userID].UnmuteDate = t.Format("2006-01-02 15:04:05")
		misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	}
	_ = misc.PunishedUsersWrite(misc.GuildMap[m.GuildID].PunishedUsers, m.GuildID)
	misc.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("__%v#%v__ has been unmuted.", user.User.Username, user.User.Discriminator))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Sends an embed message to bot-log if possible
	misc.MapMutex.Lock()
	if _, ok := misc.GuildMap[m.GuildID].MemberInfoMap[userID]; ok {
		_ = misc.UnmuteEmbed(s, misc.GuildMap[m.GuildID].MemberInfoMap[userID], m.Author.Username, guildBotLog)
	}
	misc.MapMutex.Unlock()
}

func init() {
	add(&command{
		execute:  unmuteCommand,
		trigger:  "unmute",
		desc:     "Unmutes a user",
		elevated: true,
		category: "moderation",
	})
}
