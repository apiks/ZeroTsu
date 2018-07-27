package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Removes a role from the user that uses this command if the role is between opt-in dummy roles
func leaveCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		roleID         string
		name           string

		hasRoleAlready bool
		roleExists	   bool
	)

	// Pulls info on message author
	mem, err := s.State.Member(config.ServerID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, m.Author.ID)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	// Puts the command to lowercase
	messageLowercase := strings.ToLower(m.Content)

	// Separates every word in messageLowercase and puts it in a slice
	commandStrings := strings.Split(messageLowercase, " ")
	if len(commandStrings) == 1 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "leave [channel]`")
		if err != nil {
			fmt.Println("Error:", err)
		}
		return
	}

	// Pulls the role name from strings after "leavechannel " or "leave "
	if strings.HasPrefix(messageLowercase, config.BotPrefix+"leavechannel ") {

		name = strings.Replace(messageLowercase, config.BotPrefix+"leavechannel ", "", -1)
	} else {

		name = strings.Replace(messageLowercase, config.BotPrefix+"leave ", "", -1)
	}

	// Pulls info on server roles
	deb, err := s.GuildRoles(config.ServerID)
	if err != nil {

		fmt.Println("Error:", err)
	}

	// Pulls info on server channels
	cha, err := s.GuildChannels(config.ServerID)
	if err != nil {

		fmt.Println("Error:", err)
	}

	// Checks if there's a # before the channel name and removes it if so
	if strings.Contains(name, "#") {

		name = strings.Replace(name, "#", "", -1)

		// Checks if it's in a mention format. If so then user already has access to channel
		if strings.Contains(name, "<") {

			// Fetches mention
			name = strings.Replace(name, ">", "", -1)
			name = strings.Replace(name, "<", "", -1)
			name = misc.ChMentionID(name)

			// Sends error message to user in DMs
			dm, err := s.UserChannelCreate(m.Author.ID)
			if err != nil {

				return
			}
			_, err = s.ChannelMessageSend(dm.ID, "You cannot leave "+name + " using this command.")
			return
		}
	}

	// Checks if the role exists on the server, sends error message if not
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == name {

			roleID = deb[i].ID

			if strings.Contains(deb[i].ID, roleID) {

				roleExists = true
			}
		}
	}
	if roleExists == false {

		// Sends error message to user in DMs
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {

			return
		}
		_, err = s.ChannelMessageSend(dm.ID, "There's no #"+name+", silly")
		return
	}

	// Sets role ID
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == name && roleID != "" {

			roleID = deb[i].ID
			break
		}
	}

	// Checks if the user already has the role. Sends error message if he does
	for i := 0; i < len(mem.Roles); i++ {
		if strings.Contains(mem.Roles[i], roleID) {

			hasRoleAlready = true
		}
	}
	if hasRoleAlready == false {

		var chanMention string

		// Sets the channel mention to the variable chanMention
		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				chanMention = misc.ChMention(cha[j])
			}
		}

		// Sends error message to user in DMs
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {

			return
		}
		_, err = s.ChannelMessageSend(dm.ID, "You're already out of " + chanMention + ", daaarling~")
		return
	}

	// Updates the position of opt-in-under and opt-in-above position
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == config.OptInUnder {

			misc.OptinUnderPosition = deb[i].Position
		}
		if deb[i].Name == config.OptInAbove {

			misc.OptinAbovePosition = deb[i].Position
		}
	}

	// Sets role
	role, err := s.State.Role(config.ServerID, roleID)
	if err != nil {

		fmt.Println("Error:", err)
	}

	// Removes role from user if the role is between dummy opt-ins
	if role.Position < misc.OptinUnderPosition &&
		role.Position > misc.OptinAbovePosition {

		var (
			chanMention string
		)

		err = s.GuildMemberRoleRemove(config.ServerID, m.Author.ID, roleID)
		if err != nil {

			fmt.Println("Error:", err)
		}

		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				// Sets the channel mention to the variable chanMention
				chanMention = misc.ChMention(cha[j])
			}
		}

		// Sends success message to user in DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {

			return
		}
		_, err = s.ChannelMessageSend(dm.ID, "You have left " + chanMention)
	}
}

func init() {
	add(&command{
		execute:  leaveCommand,
		trigger:  "leave",
		aliases:  []string{"leavechannel"},
		desc:     "Leave a spoiler channel.",
		deleteAfter: true,
	})
}