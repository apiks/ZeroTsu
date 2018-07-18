package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

//Sorts all spoiler roles created with the create command between the two opt-in dummy roles alphabetically
func SortRolesHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

				//Checks if the prefix + "sortroles" specifically is said
				if strings.HasPrefix(messageLowercase, config.BotPrefix+"sortroles") {

					//Initializes needed variables
					var (
						spoilerRoles      []*discordgo.Role
						underSpoilerRoles []*discordgo.Role
						rolesOrdered      []*discordgo.Role

						controlNum int
					)

					//Reads all spoiler roles created with create command from spoilerRoles.json
					misc.SpoilerRolesRead()

					if misc.SpoilerMap != nil {

						//Fetches info from the server and puts it in debPre
						debPre, err := s.GuildRoles(config.ServerID)
						if err != nil {

							fmt.Println("Error: ", err)
						}

						//Refreshes the positions of all roles in the server (because when created roles start at 0)
						for i := 0; i < len(debPre); i++ {

							spoilerRoles = append(spoilerRoles, debPre[i])
						}

						//Pushes the refreshed positions to the server
						s.GuildRoleReorder(config.ServerID, spoilerRoles)

						time.Sleep(time.Millisecond * 333)

						//Resets the value of spoilerRoles
						spoilerRoles = nil

						//Fetches the refreshed info from the server and puts it in deb
						deb, err := s.GuildRoles(config.ServerID)
						if err != nil {

							fmt.Println("Error: ", err)
						}

						//Saves the original opt-in-above position
						for i := 0; i < len(deb); i++ {
							if deb[i].Name == config.OptInAbove {

								misc.OptinAbovePosition = deb[i].Position
							}
						}

						//Adds all spoiler roles in SpoilerMap in the spoilerRoles slice
						//Adds all non-spoiler roles under opt-in-above (including it) in the underSpoilerRoles slice
						for i := 0; i < len(deb); i++ {
							_, ok := misc.SpoilerMap[deb[i].ID]
							if ok == true {

								spoilerRoles = append(spoilerRoles, misc.SpoilerMap[deb[i].ID])

								if deb[i].Position < misc.OptinAbovePosition {

									controlNum++
								}
							} else if ok == false && deb[i].Position <= misc.OptinAbovePosition && deb[i].ID != config.ServerID {

								//fmt.Println(deb[i].Name)

								underSpoilerRoles = append(underSpoilerRoles, deb[i])
							}
						}

						//If there are spoiler roles under opt-in-above it goes in to move and sort
						if controlNum > 0 {

							//Sorts the spoilerRoles slice (all spoiler roles) alphabetically
							sort.Sort(misc.SortRoleByAlphabet(spoilerRoles))

							//Moves the sorted spoiler roles above opt-in-above
							for i := 0; i < len(spoilerRoles); i++ {

								spoilerRoles[i].Position = misc.OptinAbovePosition
							}

							//Moves every non-spoiler role below opt-in-above (including it) down an amount equal to the amount of roles in the
							//spoilerRoles slice that are below opt-in-above
							for i := 0; i < len(underSpoilerRoles); i++ {

								underSpoilerRoles[i].Position = underSpoilerRoles[i].Position - controlNum
							}

							//Concatenates the two ordered role slices
							rolesOrdered = append(spoilerRoles, underSpoilerRoles...)

							//Pushes the ordered role list to the server
							s.GuildRoleReorder(config.ServerID, rolesOrdered)

							time.Sleep(time.Millisecond * 333)

							//Fetches info from the server and puts it in debPost
							debPost, err := s.GuildRoles(config.ServerID)
							if err != nil {

								fmt.Println("Error: ", err)
							}

							//Saves the new opt-in-above position
							for i := 0; i < len(debPost); i++ {
								if deb[i].Name == config.OptInAbove {

									misc.OptinAbovePosition = deb[i].Position
								}
							}

							for i := 0; i < len(spoilerRoles); i++ {

								spoilerRoles[i].Position = misc.OptinAbovePosition + len(spoilerRoles) - i
								misc.SpoilerMap[spoilerRoles[i].ID].Position = spoilerRoles[i].Position
							}

							//Pushes the sorted list to the server
							s.GuildRoleReorder(config.ServerID, spoilerRoles)

							time.Sleep(time.Millisecond * 333)

							if m.Author.ID == config.BotID {

							} else {

								//Prints success
								success := "Roles sorted."
								s.ChannelMessageSend(m.ChannelID, success)
							}
						}
					}
				}
			}
		}
	}
}
