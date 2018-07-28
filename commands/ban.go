package commands

import (
	"strings"
	"time"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Bans a user for a set period with a reason
func banCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		userID  string
		length  string
		reason  string
		success string

		validSlice bool

		temp misc.BannedUsers
	)

	commandStrings := strings.SplitN(m.Content, " ", 4)

	if len(commandStrings) != 4 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "ban [@user or userID] [time] [reason]` format. \n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes. Use 0d for permanent.")
		if err != nil {

			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
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
	length = commandStrings[2]
	reason = commandStrings[3]

	// Checks if a number is contained in length var. Fixes some cases of invalid length
	lengthSlice := strings.Split(length, "")
	for i := 0; i < len(lengthSlice); i++ {

		if _, err := strconv.ParseInt(lengthSlice[i], 10, 64); err == nil {

			validSlice = true
			break
		}
	}
	if validSlice == false {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid length. \n Usage: `" + config.BotPrefix + "ban [@user or userID] [time] [reason]` format. \n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes. Use 0d for permanent.")
		if err != nil {

			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {

				return
			}
			return
		}
		return
	}

	// Fetches user
	mem, err := s.User(userID)
	if err != nil {

		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Checks if user is in memberInfo and handles them
	if misc.MemberInfoMap == nil || misc.MemberInfoMap[userID] == nil {

		// Pulls info on user if they're in the server
		userMem, err := s.State.Member(config.ServerID, mem.ID)
		if err != nil {
			userMem, err = s.GuildMember(config.ServerID, mem.ID)
			if err != nil {
				_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in server _and_ memberInfo. Cannot ban user until they join the server.")
				if err != nil {

					_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
					if err != nil {

						return
					}
					return
				}
				return
			}
		}

		// Initializes user if he doesn't exist in memberInfo but is in server
		misc.InitializeUser(userMem)
	}

	misc.MapMutex.Lock()

	// Adds unban date to memberInfo and checks if perma
	misc.MemberInfoMap[userID].Bans = append(misc.MemberInfoMap[userID].Bans, reason)
	UnbanDate, perma := misc.ResolveTimeFromString(length)
	if perma == false {
		misc.MemberInfoMap[userID].UnbanDate = UnbanDate.Format("2006-01-02 15:04:05")

	} else {

		misc.MemberInfoMap[userID].UnbanDate = "_Never_"
	}

	misc.MapMutex.Unlock()

	// Writes to memberInfo.json
	misc.MemberInfoWrite(misc.MemberInfoMap)

	// Saves the details in temp
	temp.ID = userID
	temp.User = mem.Username

	if perma == false {

		temp.UnbanDate = UnbanDate
	} else {

		temp.UnbanDate = time.Date(9999, 9, 9, 9, 9, 9, 9, time.Local)
	}

	// Adds the now banned user to BannedUsersSlice
	misc.BannedUsersSlice = append(misc.BannedUsersSlice, temp)

	// Writes the new bannedUsers.json to file
	misc.BannedUsersWrite(misc.BannedUsersSlice)

	// Pulls the guild Name
	guild, err := s.Guild(config.ServerID)
	if err != nil {

		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Assigns success print string for user
	if perma == false {

		success = "You have been banned from " + guild.Name + ": **" + reason + "**\n\nUntil: _" + UnbanDate.Format("2006-01-02 15:04:05") + "_"
	} else {

		success = "You have been banned from " + guild.Name + ": **" + reason + "**\n\nUntil: _Forever_ \n\nIf you would like to appeal, use modmail at <https://reddit.com/r/anime>"
	}

	// Sends success string to user in DMs if able
	dm, _ := s.UserChannelCreate(userID)
	_, _ = s.ChannelMessageSend(dm.ID, success)

	// Bans the user
	err = s.GuildBanCreateWithReason(config.ServerID, mem.ID, reason, 0)
	if err != nil {

		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Sends embed bot-log message
	BanEmbed(s, m, mem, reason, length)

	// Sends a message to bot-log regarding ban
	if perma == false {

		_, err = s.ChannelMessageSend(m.ChannelID, mem.Username + "#" + mem.Discriminator + " has been banned by "+
			m.Author.Username+ " until _"+ UnbanDate.Format("2006-01-02 15:04:05")+ "_")
		if err != nil {

			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {

				return
			}
			return
		}
	} else {

		_, err = s.ChannelMessageSend(m.ChannelID, mem.Username + "#" + mem.Discriminator + " has been permabanned by "+
			m.Author.Username)
		if err != nil {

			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {

				return
			}
			return
		}
	}
}

func BanEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.User, reason string, length string) error {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail

		// Embed slice and its fields
		embedField         []*discordgo.MessageEmbedField
		embedFieldUserID   discordgo.MessageEmbedField
		embedFieldReason   discordgo.MessageEmbedField
		embedFieldDuration discordgo.MessageEmbedField
	)

	// Saves user avatar as thumbnail
	embedThumbnail.URL = mem.AvatarURL("128")

	// Sets field titles
	embedFieldUserID.Name = "User ID:"
	embedFieldReason.Name = "Reason:"
	embedFieldDuration.Name = "Duration:"

	// Sets field content
	embedFieldUserID.Value = mem.ID
	embedFieldReason.Value = reason
	embedFieldDuration.Value = length

	// Sets field inline
	embedFieldUserID.Inline = true
	embedFieldReason.Inline = true
	embedFieldDuration.Inline = true

	// Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldUserID)
	embedField = append(embedField, &embedFieldDuration)
	embedField = append(embedField, &embedFieldReason)

	// Sets embed title and its description (which it uses the same way as a field)
	embedMess.Title = mem.Username + "#" + mem.Discriminator + " was banned by " + m.Author.Username

	// Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	// Sends embed in bot-log channel
	_, err := s.ChannelMessageSendEmbed(config.BotLogID, &embedMess)
	return err
}

//func init() {
//	add(&command{
//		execute:  banCommand,
//		trigger:  "ban",
//		desc:     "Bans a user for a set period of time.",
//		elevated: true,
//	})
//}