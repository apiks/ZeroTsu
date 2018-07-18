package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

//Sorts all channels in a given category alphabetically
func SortCategoryHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

			//Checks for mod permissions
			if misc.HasPermissions(mem) {

				//Initializes needed variables
				var (
					categoryID       string
					categoryPosition int
					categoryChannels []*discordgo.Channel
				)

				//Checks if the prefix + "sortcategory" specifically is said
				if strings.HasPrefix(messageLowercase, config.BotPrefix+"sortcategory") {

					//Pulls the user from strings after "sortcategory "
					categoryString := strings.Replace(messageLowercase, config.BotPrefix+"sortcategory ", "", 1)

					//Fetches channel info from the server and puts it in deb
					deb, err := s.GuildChannels(config.ServerID)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					for i := 0; i < len(deb); i++ {

						//Puts channel name to lowercase
						nameLowercase := strings.ToLower(deb[i].Name)

						//Compares if the categoryString is either a valid category name or ID
						if nameLowercase == categoryString || deb[i].ID == categoryString {
							if deb[i].Type == discordgo.ChannelTypeGuildCategory {

								categoryID = deb[i].ID

								categoryPosition = deb[i].Position
							}
						}
					}

					if categoryID != "" {

						for i := 0; i < len(deb); i++ {
							if deb[i].ParentID == categoryID {

								categoryChannels = append(categoryChannels, deb[i])
							}
						}

						//Sorts the categoryChannels slice (all channels in the inputted category) alphabetically
						sort.Sort(misc.SortChannelByAlphabet(categoryChannels))

						for i := 0; i < len(categoryChannels); i++ {

							categoryChannels[i].Position = categoryPosition + i + 1
						}

						//Pushes the sorted list to the server
						err = s.GuildChannelsReorder(config.ServerID, categoryChannels)
						if err != nil {

							fmt.Println("Error: ", err)
						}

						if m.Author.ID == config.BotID {

						} else {

							//Prints success
							success := "Category `" + categoryString + "` sorted"
							s.ChannelMessageSend(m.ChannelID, success)
						}
					}
				}
			}
		}
	}
}
