package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Unbans a user and updates their memberInfo entry
func unbanCommand(s *discordgo.Session, m *discordgo.Message) {

	var banFlag = false

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"unban [@user, userID, or username#discrim]` format.\n\n"+
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

	user, err := s.User(userID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Goes through every banned user from BannedUsersSlice and if the user is in it, confirms that user is a temp ban
	functionality.MapMutex.Lock()
	if len(functionality.GuildMap[m.GuildID].PunishedUsers) == 0 {
		_, err = s.ChannelMessageSend(m.ChannelID, "No bans found.")
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
			banFlag = true

			// Removes the ban from punishedUsers
			functionality.GuildMap[m.GuildID].PunishedUsers = append(functionality.GuildMap[m.GuildID].PunishedUsers[:i], functionality.GuildMap[m.GuildID].PunishedUsers[i+1:]...)
			break
		}
	}
	functionality.MapMutex.Unlock()

	// Check if the user is banned using other means
	if !banFlag {
		bans, err := s.GuildBans(m.GuildID)
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("__%v#%v__ is not banned.", user.Username, user.Discriminator))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Removes the ban
	err = s.GuildBanDelete(m.GuildID, userID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Saves time of unban command usage
	t := time.Now()

	// Updates unban date in memberInfo.json entry if possible and writes to storage
	functionality.MapMutex.Lock()
	if _, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]; ok {
		functionality.GuildMap[m.GuildID].MemberInfoMap[userID].UnbanDate = t.Format("2006-01-02 15:04:05")
		functionality.WriteMemberInfo(functionality.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	}
	_ = functionality.PunishedUsersWrite(functionality.GuildMap[m.GuildID].PunishedUsers, m.GuildID)
	functionality.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("__%v#%v__ has been unbanned.", user.Username, user.Discriminator))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Sends an embed message to bot-log if possible
	functionality.MapMutex.Lock()
	if _, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]; ok {
		if guildSettings.BotLog != nil {
			if guildSettings.BotLog.ID != "" {
				_ = functionality.UnbanEmbed(s, functionality.GuildMap[m.GuildID].MemberInfoMap[userID], m.Author.Username, guildSettings.BotLog.ID)
			}
		}
	}
	functionality.MapMutex.Unlock()
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    unbanCommand,
		Trigger:    "unban",
		Desc:       "Unbans a user",
		Permission: functionality.Mod,
		Module:     "moderation",
	})
}
