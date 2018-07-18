package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

//Removes role from user that uses this command if the role is between opt-in dummy roles
func LeaveChannelHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Checks if it's within the /r/anime server
	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		ch, err = s.Channel(m.ChannelID)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
	if ch.GuildID == config.ServerID {

		if strings.HasPrefix(m.Content, config.BotPrefix) {

			mem, err := s.State.Member(config.ServerID, m.Author.ID)
			if err != nil {
				mem, err = s.GuildMember(config.ServerID, m.Author.ID)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}

			//Puts the command to lowercase
			messageLowercase := strings.ToLower(m.Content)

			if (strings.HasPrefix(messageLowercase, config.BotPrefix+"leavechannel ") && (messageLowercase != (config.BotPrefix + "leavechannel"))) ||
				(strings.HasPrefix(messageLowercase, config.BotPrefix+"leave ") && (messageLowercase != (config.BotPrefix + "leave"))) {

				if m.Author.ID == config.BotID {
					return
				}

				//Initializes needed variables
				var (
					roleID         string
					hasRoleAlready = false
					roleExists     = false
					name           string
				)

				//Deletes the message that was sent so it doesn't clog up the channel.
				s.ChannelMessageDelete(m.ChannelID, m.ID)

				//Pulls the name from strings after "leavechannel " or "leave "
				if strings.Contains(messageLowercase, config.BotPrefix+"leavechannel ") {

					name = strings.Replace(messageLowercase, config.BotPrefix+"leavechannel ", "", -1)
				} else {

					name = strings.Replace(messageLowercase, config.BotPrefix+"leave ", "", -1)
				}

				//Pulls info on server roles
				deb, err := s.GuildRoles(config.ServerID)
				if err != nil {

					fmt.Println("Error: ", err)
				}

				//Pulls info on server channels
				cha, err := s.GuildChannels(config.ServerID)
				if err != nil {

					fmt.Println("Error: ", err)
				}

				//Checks if there's a # before the channel name and removes it
				if strings.Contains(name, "#") {

					name = strings.Replace(name, "#", "", -1)

					//Checks if it's a channel mention and if it is, removes the arrows that come with it
					//Also makes bools into true since it'd exist if in mention form and user has it
					if strings.Contains(name, "<") {

						roleID = strings.Replace(name, ">", "", -1)
						roleID = strings.Replace(roleID, "<", "", -1)

						//Fixes name since the mention shows only channel ID
						for i := 0; i < len(cha); i++ {
							if cha[i].ID == roleID {

								name = cha[i].Name
								break
							}
						}
					}
				}

				//Sets role ID
				for i := 0; i < len(deb); i++ {
					if deb[i].Name == name {

						roleID = deb[i].ID
						break
					}
				}

				roleName := name //Fixes naming issue

				//Assigns the position of opt-in-under position
				for i := 0; i < len(deb); i++ {
					if deb[i].Name == config.OptInUnder {

						misc.OptinUnderPosition = deb[i].Position
					}
				}

				//Assigns the position of opt-in-above position
				for i := 0; i < len(deb); i++ {
					if deb[i].Name == config.OptInAbove {

						misc.OptinAbovePosition = deb[i].Position
					}
				}

				//Checks if the user has the role already
				for i := 0; i < len(mem.Roles); i++ {

					fmt.Println(mem.Roles[i])

					if strings.Contains(mem.Roles[i], roleID) && roleID != "" {

						fmt.Println("It's in")

						hasRoleAlready = true
					}
				}

				//Checks if the role exists on the server
				for i := 0; i < len(deb); i++ {
					if deb[i].Name == roleName {

						roleID = deb[i].ID
						if strings.Contains(deb[i].ID, roleID) {

							roleExists = true
						}
					}
				}

				//Checks if the role is above opt-in-above and under opt-in-under and then checks if it exists and if the user
				//already has it. If everything is true it removes the role from him
				for i := 0; i < len(deb); i++ {
					if deb[i].Name == roleName &&
						hasRoleAlready == true &&
						roleExists == true &&
						deb[i].Position < misc.OptinUnderPosition &&
						deb[i].Position > misc.OptinAbovePosition {

						roleID = deb[i].ID
						s.GuildMemberRoleRemove(config.ServerID, m.Author.ID, roleID)

						success := "You have left #" + roleName

						//Creates a DM connection and assigns it to the variable dm
						dm, err := s.UserChannelCreate(m.Author.ID)
						if err != nil {

							fmt.Println("Error: ", err)
						}

						//Sends a message to that DM connection
						_, err = s.ChannelMessageSend(dm.ID, success)
						if err != nil {

							fmt.Println("Error: ", err)
						}
					}
				}

				//If the role exists and the user already has it it prints that they already have it.
				//If the role doesn't exist it tells them so.
				if hasRoleAlready == false &&
					roleExists == true {

					failure := "You're already out of #" + roleName + ", daaarling~"

					//Creates a DM connection and assigns it to dm
					dm, err := s.UserChannelCreate(m.Author.ID)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Sends a message to that DM connection
					_, err = s.ChannelMessageSend(dm.ID, failure)
					if err != nil {

						fmt.Println("Error: ", err)
					}
				} else if roleExists == false {

					failure := "There's no #" + roleName + ", silly"

					//Creates a DM connection and assigns it to dm
					dm, err := s.UserChannelCreate(m.Author.ID)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Sends a message to that DM connection
					_, err = s.ChannelMessageSend(dm.ID, failure)
					if err != nil {

						fmt.Println("Error: ", err)
					}
				}
			}
		}
	}
}
