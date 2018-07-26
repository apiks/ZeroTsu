package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

//Adds role to user that uses this command if the role is between opt-in dummy roles
func JoinChannelHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

			//Pulls info on message author
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

			if (strings.HasPrefix(messageLowercase, config.BotPrefix+"joinchannel ") && (messageLowercase != (config.BotPrefix + "joinchannel"))) ||
				(strings.HasPrefix(messageLowercase, config.BotPrefix+"join ") && (messageLowercase != (config.BotPrefix + "join"))) {

				if m.Author.ID == config.BotID {
					return
				}

				//Initializes necessary variables
				var (
					roleID         string
					hasRoleAlready = false
					roleExists     = false
					topic          string
					chanMention    string
					name           string
				)

				//Deletes the message that was sent so it doesn't clog up the channel.
				s.ChannelMessageDelete(m.ChannelID, m.ID)

				//Pulls the name from strings after "joinchannel " or "join "
				if strings.HasPrefix(messageLowercase, config.BotPrefix+"joinchannel ") {

					name = strings.Replace(messageLowercase, config.BotPrefix+"joinchannel ", "", -1)
				} else {

					name = strings.Replace(messageLowercase, config.BotPrefix+"join ", "", -1)
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
					if deb[i].Name == name && roleID != "" {

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

				//Checks if the role exists on the server
				for i := 0; i < len(deb); i++ {
					if deb[i].Name == roleName {

						roleID = deb[i].ID

						if strings.Contains(deb[i].ID, roleID) {

							roleExists = true
						}
					}
				}

				//Checks if the user has the role already
				for i := 0; i < len(mem.Roles); i++ {
					if strings.Contains(mem.Roles[i], roleID) {

						hasRoleAlready = true
					}
				}

				//Checks if the role is above opt-in-above and under opt-in-under and then checks if it exists and if the user
				//already has it. If everything is true it assigns him the role
				for i := 0; i < len(deb); i++ {
					if deb[i].Name == roleName &&
						hasRoleAlready == false &&
						roleExists == true &&
						deb[i].Position < misc.OptinUnderPosition &&
						deb[i].Position > misc.OptinAbovePosition {

						roleID = deb[i].ID
						s.GuildMemberRoleAdd(config.ServerID, m.Author.ID, roleID)

						for j := 0; j < len(cha); j++ {
							if cha[j].Name == roleName {

								topic = cha[j].Topic

								//Assigns the channel mention to the variable chanMention
								chanMention = misc.ChMention(cha[j])
							}
						}

						success := "You have joined " + chanMention

						if topic != "" {

							success = success + "\n **Topic:** " + topic
						}

						//Creates a DM connection and assigns it to the variable dm
						dm, err := s.UserChannelCreate(m.Author.ID)
						if err != nil {

							fmt.Println("Error: ", err)
						}

						if dm != nil {

							//Sends a message to that DM connection
							_, err = s.ChannelMessageSend(dm.ID, success)
							if err != nil {

								fmt.Println("Error: ", err)
							}
						}
					}
				}

				//If the role exists and the user already has it it prints that they already have it.
				//If the role doesn't exist it tells them so.
				if hasRoleAlready == true &&
					roleExists == true {

					for j := 0; j < len(cha); j++ {
						if cha[j].Name == roleName || cha[j].ID == roleID {

							topic = cha[j].Topic

							//Assigns the channel mention to the variable chanMention
							chanMention = misc.ChMention(cha[j])
						}
					}

					failure := "You're already in " + chanMention + ", daaarling~"

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
