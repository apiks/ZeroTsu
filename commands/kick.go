package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Kicks a user from the server with a reason
func kickCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		userID string
		reason string
	)

	// Pulls the user and and reason from message
	commandStrings := strings.SplitN(m.Content, " ", 4)

	// Checks if it has all parameters, else error
	if len(commandStrings) == 3 {

		userID = misc.GetUserID(s, m, commandStrings)
		reason = commandStrings[2]

	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Please use `"+config.BotPrefix+"kick [@user or userID] [reason]` format.")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}

	// Fetches user from server
	mem, err := s.User(userID)
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid user. Please use `"+config.BotPrefix+"kick [@user or userID] [reason]` format.")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}

	// Pulls info on user
	userMem, err := s.State.Member(config.ServerID, mem.ID)
	if err != nil {
		userMem, err = s.GuildMember(config.ServerID, mem.ID)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	// Checks if user is in server
	if userMem == nil {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in server. Cannot kick user until they rejoin the server.")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}

	// Initialize user if they are not in memberInfo
	if misc.MemberInfoMap == nil || misc.MemberInfoMap[userID] == nil {

		misc.InitializeUser(userMem)
	}

	// Adds kick reason to user memberInfo info
	misc.MapMutex.Lock()
	misc.MemberInfoMap[userID].Kicks = append(misc.MemberInfoMap[userID].Kicks, reason)
	misc.MapMutex.Unlock()

	// Writes memberInfo.json
	misc.MemberInfoWrite(misc.MemberInfoMap)

	// Fetches the guild Name
	guild, err := s.Guild(config.ServerID)
	if err != nil {

		fmt.Println("Error:", err)
	}
	guildName := guild.Name

	// Sends message to user DMs if able
	dm, err := s.UserChannelCreate(mem.ID)
	_, err = s.ChannelMessageSend(dm.ID, "You have been kicked from " + guildName + ":\n**" + reason + "**")

	// Kicks the user from the server with a reason
	err = s.GuildMemberDeleteWithReason(config.ServerID, mem.ID, reason)
	if err != nil {

		fmt.Println("Error:", err)
	}

	//Sends embed bot-log message
	KickEmbed(s, m, mem, reason)
}

func KickEmbed(s *discordgo.Session, m *discordgo.Message, mem *discordgo.User, reason string) {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail

		// Embed slice and its fields
		embedField       []*discordgo.MessageEmbedField
		embedFieldUserID discordgo.MessageEmbedField
		embedFieldReason discordgo.MessageEmbedField
	)

	// Saves user avatar as thumbnail
	embedThumbnail.URL = mem.AvatarURL("128")

	// Sets field titles
	embedFieldUserID.Name = "User ID:"
	embedFieldReason.Name = "Reason:"

	// Sets field content
	embedFieldUserID.Value = mem.ID
	embedFieldReason.Value = reason

	// Sets field inline
	embedFieldUserID.Inline = true
	embedFieldReason.Inline = true

	// Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldUserID)
	embedField = append(embedField, &embedFieldReason)

	// Sets embed title and its description (which it uses the same way as a field)
	embedMess.Title = mem.Username + "#" + mem.Discriminator + " was kicked by " + m.Author.Username

	// Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	// Sends embed in bot-log channel
	_, err := s.ChannelMessageSendEmbed(config.BotLogID, &embedMess)
	if err != nil {

		fmt.Println("Error:", err)
	}
}

//func init() {
//	add(&command{
//		execute:  kickCommand,
//		trigger:  "kick",
//		desc:     "Kicks a user from the server and logs reason.",
//		elevated: true,
//	})
//}