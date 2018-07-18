package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

//Sends a user avatar as a message
func AvatarHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	if strings.HasPrefix(m.Content, config.BotPrefix) {

		//Puts the command to lowercase
		messageLowercase := strings.ToLower(m.Content)

		//Checks if BotPrefix + avatar was used
		if strings.HasPrefix(messageLowercase, config.BotPrefix+"avatar ") && (messageLowercase != (config.BotPrefix + "avatar")) {

			if m.Author.ID == config.BotID {
				return
			}

			//Pulls the user from strings after "avatar "
			user := strings.Replace(messageLowercase, config.BotPrefix+"avatar ", "", 1)

			//Checks if it's an @ mention or just user ID. If the former it removes fluff
			if strings.Contains(user, "@") {

				//Removes "<@", "!" and ">" which is what @ mention returns above
				user = strings.Replace(user, "<@", "", -1)
				user = strings.Replace(user, ">", "", -1)
				user = strings.Replace(user, "!", "", -1)

			}

			//Fetches user from server
			mem, err := s.User(user)
			if err != nil {

				fmt.Println("Error: ", err)

				//Sends a message to the channel with error message
				_, err = s.ChannelMessageSend(m.ChannelID, user+" is not a valid user. Use `@user` or `user ID`.")
				if err != nil {

					fmt.Println("Error: ", err)
				}
			} else if err == nil {

				//Saves the avatar URL to avatar variable with image size 256 (how big image is on screen)
				avatar := mem.AvatarURL("256")

				//Sends a message to the channel with the avatar URL
				_, err = s.ChannelMessageSend(m.ChannelID, avatar)
				if err != nil {

					fmt.Println("Error: ", err)
				}
			}
		}
	}
}
