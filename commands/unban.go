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

	// Puts entire message in lowercase
	messageLowercase := strings.ToLower(m.Content)

	// Separates every word in the message and puts it in a slice
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Wrong amount of parameters. Please use `"+config.BotPrefix+"unban [@user or userID]` format.")
		if err != nil {

			fmt.Println("Error:", err)
		}
	}

	userID := misc.GetUserID(s, m, commandStrings)
	user, err := s.User(userID)
	if err != nil {

		fmt.Println("Error:", err)
	}

	var banFlag = false

	// Goes through every banned user from bannedUsers.json and if the user is in it, confirms that user is a banned one
	if misc.BannedUsersSlice == nil {

		_, err = s.ChannelMessageSend(m.ChannelID, "No bans in storage.")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}

	for i := 0; i < len(misc.BannedUsersSlice); i++ {
		if misc.BannedUsersSlice[i].ID == userID {

			banFlag = true

			// Removes the ban from bannedUsers.json and writes to bannedUsers.json
			misc.BannedUsersSlice = append(misc.BannedUsersSlice[:i], misc.BannedUsersSlice[i+1:]...)
			misc.BannedUsersWrite(misc.BannedUsersSlice)
			break
		}
	}

	if banFlag == false {
		_, err := s.ChannelMessageSend(m.ChannelID, user.Username+"#"+user.Discriminator+" is not banned.")
		if err != nil {

			fmt.Println("Error:", err)
		}
	} else {

		// Removes the ban
		err = s.GuildBanDelete(config.ServerID, userID)
		if err != nil {

			fmt.Println("Error:", err)
		}

		_, err = s.ChannelMessageSend(m.ChannelID, user.Username+"#"+user.Discriminator+" has been unbanned.")
		if err != nil {

			fmt.Println("Error:", err)
		}

		// Saves time of unban command usage
		t := time.Now()

		// Updates unban date in memberInfo.json entry
		misc.MapMutex.Lock()
		misc.MemberInfoMap[userID].UnbanDate = t.Format("2006-01-02 15:04:05")
		misc.MapMutex.Unlock()

		// Writes to memberInfo.json
		misc.MemberInfoWrite(misc.MemberInfoMap)
	}
}

//func init() {
//	add(&command{
//		execute:  unbanCommand,
//		trigger:  "unban",
//		desc:     "Unbans a user.",
//		elevated: true,
//	})
//}