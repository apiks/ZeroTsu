package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

//Handler for channel lock/unlock
func ChannelLockHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

			//Checks for mod permissions
			if misc.HasPermissions(mem) {
				if m.Author.ID == config.BotID {
					return
				}

				//Puts the command to lowercase
				messageLowercase := strings.ToLower(m.Content)

				//Initializes variable to check if the locked channel is a spoiler channel
				var spoilerRole = false

				if messageLowercase == config.BotPrefix+"lock" {

					//Initializes needed variables
					var (
						roleExists       = false
						roleID           string
						roleTempPosition int
					)

					//Prints success message
					_, _ = s.ChannelMessageSend(m.ChannelID, "🔒 This channel has been locked.")

					//Pulls info on the channel the message is in
					cha, err := s.Channel(m.ChannelID)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Fetches info on server roles from the server and puts it in deb
					deb, err := s.GuildRoles(config.ServerID)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Updates opt-in-under and opt-in-above position
					for i := 0; i < len(deb); i++ {
						if deb[i].Name == config.OptInUnder {

							misc.OptinUnderPosition = deb[i].Position
						} else if deb[i].Name == config.OptInAbove {

							misc.OptinAbovePosition = deb[i].Position
						} else if deb[i].Name == cha.Name {

							roleID = deb[i].ID
							roleTempPosition = deb[i].Position
						}
					}

					//Checks if the channel being locked is between the opt-ins
					for i := 0; i < len(deb); i++ {
						if roleTempPosition < misc.OptinUnderPosition &&
							roleTempPosition > misc.OptinAbovePosition {

							spoilerRole = true
						}
					}

					if spoilerRole == true {

						//Removes send permissions only from the channel role if it's a spoiler channel
						err = s.ChannelPermissionSet(m.ChannelID, roleID, "role", discordgo.PermissionReadMessages, discordgo.PermissionSendMessages)
						if err != nil {

							fmt.Println("Error: ", err)
						}
					} else {

						//Removes send permission from @everyone
						err = s.ChannelPermissionSet(m.ChannelID, config.ServerID, "role", 0, discordgo.PermissionSendMessages)
						if err != nil {

							fmt.Println("Error: ", err)
						}
					}

					//Checks if the channel has a permission overwrite for mods
					for i := 0; i < len(cha.PermissionOverwrites); i++ {
						for _, goodRole := range config.CommandRoles {
							if cha.PermissionOverwrites[i].ID == goodRole {

								roleExists = true
							}
						}
					}

					//If the mod permission overwrite doesn't exist it adds it
					if roleExists == false {
						for i := 0; i < len(deb); i++ {
							for _, goodRole := range config.CommandRoles {

								s.ChannelPermissionSet(m.ChannelID, goodRole, "role", discordgo.PermissionAll, 0)
							}
						}
					}

					//Prints success message in bot log
					s.ChannelMessageSend(config.BotLogID, "🔒 "+chMention(cha)+" was locked by "+m.Author.Username)

				} else if messageLowercase == config.BotPrefix+"unlock" {

					//Initializes needed variables
					var (
						def              int
						roleID           string
						roleTempPosition int
					)

					//Sets permission variable to be neutral for send messages
					def &= ^discordgo.PermissionSendMessages

					//Pulls info on the channel the message is in
					cha, err := s.Channel(m.ChannelID)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Fetches info on server roles from the server and puts it in deb
					deb, err := s.GuildRoles(config.ServerID)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Updates opt-in-under and opt-in-above position
					for i := 0; i < len(deb); i++ {
						if deb[i].Name == config.OptInUnder {

							misc.OptinUnderPosition = deb[i].Position
						} else if deb[i].Name == config.OptInAbove {

							misc.OptinAbovePosition = deb[i].Position
						} else if deb[i].Name == cha.Name {

							roleID = deb[i].ID
							roleTempPosition = deb[i].Position
						}
					}

					//Checks if the channel being locked is between the opt-ins
					for i := 0; i < len(deb); i++ {
						if roleTempPosition < misc.OptinUnderPosition &&
							roleTempPosition > misc.OptinAbovePosition {

							spoilerRole = true
						}
					}

					if spoilerRole == true {

						//Adds send permissions only to the channel role if it's a spoiler channel
						err = s.ChannelPermissionSet(m.ChannelID, roleID, "role", misc.SpoilerPerms, 0)
						if err != nil {

							fmt.Println("Error: ", err)
						}
					} else {

						//Adds send permission from @everyone
						s.ChannelPermissionSet(m.ChannelID, config.ServerID, "role", def, 0)
					}

					//Prints success message
					_, _ = s.ChannelMessageSend(m.ChannelID, "🔓 This channel has been unlocked.")

					//Prints success message in bot log
					_, _ = s.ChannelMessageSend(config.BotLogID, "🔓 "+chMention(cha)+" was unlocked by "+m.Author.Username)
				}
			}
		}
	}
}

func chMention(ch *discordgo.Channel) string {
	return fmt.Sprintf("<#%s>", ch.ID)
}
