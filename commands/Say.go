package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

//Sends a message from the bot to the channel
func SayHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

		//Checks for permissions and command
		if misc.HasPermissions(mem) {
			if strings.HasPrefix(messageLowercase, config.BotPrefix+"say ") && (messageLowercase != (config.BotPrefix + "say")) {

				if m.Author.ID == config.BotID {
					return
				}

				//Deletes the say message
				s.ChannelMessageDelete(m.ChannelID, m.ID)

				//Pulls the sentence from strings after "say "
				sentence := strings.Replace(m.Content, config.BotPrefix+"say ", "", -1)

				//Sends the sentence to the channel the original message was in.
				s.ChannelMessageSend(m.ChannelID, sentence)
			}
		}
	}
}
