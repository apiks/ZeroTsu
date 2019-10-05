package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Unbans a user and updates their memberInfo entry
func unbanCommand(s *discordgo.Session, m *discordgo.Message) {

	var banFlag = false

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"unban [@user, userID, or username#discrim]` format.\n\n"+
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

	user, err := s.User(userID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Goes through every banned user from BannedUsersSlice and if the user is in it, confirms that user is a temp ban
	misc.MapMutex.Lock()
	if len(misc.GuildMap[m.GuildID].PunishedUsers) == 0 {
		_, err = s.ChannelMessageSend(m.ChannelID, "No bans found.")
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
			banFlag = true

			// Removes the ban from punishedUsers
			misc.GuildMap[m.GuildID].PunishedUsers = append(misc.GuildMap[m.GuildID].PunishedUsers[:i], misc.GuildMap[m.GuildID].PunishedUsers[i+1:]...)
			break
		}
	}
	misc.MapMutex.Unlock()

	// Check if the user is banned using other means
	if !banFlag {
		bans, err := s.GuildBans(m.GuildID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
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
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Removes the ban
	err = s.GuildBanDelete(m.GuildID, userID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Saves time of unban command usage
	t := time.Now()

	// Updates unban date in memberInfo.json entry if possible and writes to storage
	misc.MapMutex.Lock()
	if _, ok := misc.GuildMap[m.GuildID].MemberInfoMap[userID]; ok {
		misc.GuildMap[m.GuildID].MemberInfoMap[userID].UnbanDate = t.Format("2006-01-02 15:04:05")
		misc.WriteMemberInfo(misc.GuildMap[m.GuildID].MemberInfoMap, m.GuildID)
	}
	_ = misc.PunishedUsersWrite(misc.GuildMap[m.GuildID].PunishedUsers, m.GuildID)
	misc.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("__%v#%v__ has been unbanned.", user.Username, user.Discriminator))
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
		_ = misc.UnbanEmbed(s, misc.GuildMap[m.GuildID].MemberInfoMap[userID], m.Author.Username, guildBotLog)
	}
	misc.MapMutex.Unlock()
}

func init() {
	add(&command{
		execute:  unbanCommand,
		trigger:  "unban",
		desc:     "Unbans a user",
		elevated: true,
		category: "moderation",
	})
}
