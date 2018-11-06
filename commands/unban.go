package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Unbans a user and updates their memberInfo entry
func unbanCommand(s *discordgo.Session, m *discordgo.Message) {

	var banFlag = false

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"unban [@user, userID, or username#discrim]` format.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	userID, err := misc.GetUserID(s, m, commandStrings)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	user, err := s.User(userID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Goes through every banned user from BannedUsersSlice and if the user is in it, confirms that user is a banned one
	if len(misc.BannedUsersSlice) == 0 {
		_, err = s.ChannelMessageSend(m.ChannelID, "No bans found.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	for i := 0; i < len(misc.BannedUsersSlice); i++ {
		if misc.BannedUsersSlice[i].ID == userID {
			banFlag = true

			// Removes the ban from BannedUsersSlice
			misc.BannedUsersSlice = append(misc.BannedUsersSlice[:i], misc.BannedUsersSlice[i+1:]...)
			break
		}
	}

	if !banFlag {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("__%v#%v__ is not banned.", user.Username, user.Discriminator))
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Removes the ban
	err = s.GuildBanDelete(config.ServerID, userID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Saves time of unban command usage
	t := time.Now()

	// Updates unban date in memberInfo.json entry
	misc.MapMutex.Lock()
	misc.MemberInfoMap[userID].UnbanDate = t.Format("2006-01-02 15:04:05")
	misc.MapMutex.Unlock()

	// Writes to memberInfo.json
	misc.MemberInfoWrite(misc.MemberInfoMap)

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("__%v#%v__ has been unbanned.", user.Username, user.Discriminator))
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Sends an embed message to bot-log
	misc.MapMutex.Lock()
	err = misc.UnbanEmbed(s, misc.MemberInfoMap[userID], m.Author.Username)
	misc.MapMutex.Unlock()
}

func init() {
	add(&command{
		execute:  unbanCommand,
		trigger:  "unban",
		desc:     "Unbans a user.",
		elevated: true,
		category: "punishment",
	})
}