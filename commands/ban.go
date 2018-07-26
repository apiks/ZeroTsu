package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

func banCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		userID string
		length string
		reason string
		success string

		temp    misc.BannedUsers
	)

	// Pulls the user, time and reason from messageLowercase
	commandStrings := strings.SplitN(m.Content, " ", 4)

	// Checks if it has all parameters, else error
	if len(commandStrings) == 4 {

		userID = misc.GetUserID(s, m, commandStrings)
		length = commandStrings[2]
		reason = commandStrings[3]

	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error. Please use `"+config.BotPrefix+"ban [@user or userID] [time] [reason]` format. \n\n"+
			"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes. Use 0d for permanent.")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}

	// Checks if user is in memberInfo. Prints error if not
	if misc.MemberInfoMap == nil || misc.MemberInfoMap[userID] == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: User is not in memberInfo. Please ban manually.")
		if err != nil {

			fmt.Println("Error:", err)
		}
	}

	// Fetches user
	mem, err := s.User(userID)
	if err != nil {

		fmt.Println("Error:", err)
		return
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

		fmt.Println("Error: ", err)
	}
	guildName := guild.Name

	// Assigns success print string for user
	if perma == false {

		success = "You have been banned from " + guildName + ": **" + reason + "**\n\nUntil: _" + UnbanDate.Format("2006-01-02 15:04:05") + "_"
	} else {

		success = "You have been banned from " + guildName + ": **" + reason + "**\n\nUntil: _Forever_ \n\nIf you would like to appeal, use modmail at <https://reddit.com/r/anime>"
	}

	// Sends success string to user in DMs
	dm, err := s.UserChannelCreate(userID)
	if err != nil {

		fmt.Println("Error: ", err)
	}
	_, err = s.ChannelMessageSend(dm.ID, success)
	if err != nil {

		fmt.Println("Error: ", err)
	}

	// Bans the user
	err = s.GuildBanCreateWithReason(config.ServerID, mem.ID, reason, 0)
	if err != nil {

		_, err = s.ChannelMessageSend(m.ChannelID, "Error:" + err.Error())
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}

	// Sends embed bot-log message
	BanEmbed(s, m, mem, reason, length)

	// Sends a message to bot-log regarding ban
	if perma == false {

		_, err = s.ChannelMessageSend(m.ChannelID, mem.Username+"#"+mem.Discriminator+" has been banned by "+
			m.Author.Username+" until _"+UnbanDate.Format("2006-01-02 15:04:05")+"_")
		if err != nil {

			fmt.Println("Error:", err)
		}
	} else {

		_, err = s.ChannelMessageSend(m.ChannelID, mem.Username+"#"+mem.Discriminator+" has been permabanned by "+
			m.Author.Username)
		if err != nil {

			fmt.Println("Error:", err)
		}
	}
}

func BanEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.User, reason string, length string) {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail

		//Embed slice and its fields
		embedField         []*discordgo.MessageEmbedField
		embedFieldUserID   discordgo.MessageEmbedField
		embedFieldReason   discordgo.MessageEmbedField
		embedFieldDuration discordgo.MessageEmbedField
	)

	//Saves user avatar as thumbnail
	embedThumbnail.URL = mem.AvatarURL("128")

	//Sets field titles
	embedFieldUserID.Name = "User ID:"
	embedFieldReason.Name = "Reason:"
	embedFieldDuration.Name = "Duration:"

	//Sets field content
	embedFieldUserID.Value = mem.ID
	embedFieldReason.Value = reason
	embedFieldDuration.Value = length

	//Sets field inline
	embedFieldUserID.Inline = true
	embedFieldReason.Inline = true
	embedFieldDuration.Inline = true

	//Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldUserID)
	embedField = append(embedField, &embedFieldDuration)
	embedField = append(embedField, &embedFieldReason)

	//Sets embed title and its description (which it uses the same way as a field)
	embedMess.Title = mem.Username + "#" + mem.Discriminator + " was banned by " + m.Author.Username

	//Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	//Send embed in bot-log channel
	_, err := s.ChannelMessageSendEmbed(config.BotLogID, &embedMess)
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

//func init() {
//	add(&command{
//		execute:  banCommand,
//		trigger:  "ban",
//		desc:     "Bans a user for a set period of time.",
//		elevated: true,
//	})
//}