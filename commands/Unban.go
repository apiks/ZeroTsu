package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

//Unbans banned user
func UnbanHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

			_, err := s.State.Member(config.ServerID, m.Author.ID)
			if err != nil {
				_, err = s.GuildMember(config.ServerID, m.Author.ID)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}

			//Puts the command to lowercase
			messageLowercase := strings.ToLower(m.Content)

			//Checks if BotPrefix + unban was used
			if strings.HasPrefix(messageLowercase, config.BotPrefix+"unban ") && (messageLowercase != (config.BotPrefix + "unban")) {

				if m.Author.ID == config.BotID {
					return
				}

				//Pulls the user from strings after "unban "
				user := strings.Replace(messageLowercase, config.BotPrefix+"unban ", "", 1)

				//Checks if it's an @ mention or just user ID. If the former it removes fluff
				if strings.Contains(user, "@") {

					//Removes "<@" and ">" which is what @ mention returns above
					user = strings.TrimPrefix(user, "<@")
					user = strings.TrimSuffix(user, ">")
				} else if strings.Contains(user, "!") {

					user = strings.TrimPrefix(user, "!<@")
					user = strings.TrimSuffix(user, ">")
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

					//Reads memberInfo.json
					misc.MemberInfoRead()

					//Reads bannedUsers.json
					misc.BannedUsersRead()

					//Initializes variable that knows whether user is banned in check below
					var banFlag = false

					//Goes through every banned user from bannedUsers.json and if the user is in it, confirms that user is a banned one
					if misc.BannedUsersSlice != nil {
						for i := 0; i < len(misc.BannedUsersSlice); i++ {
							if misc.BannedUsersSlice[i].ID == mem.ID {

								banFlag = true

								//Removes the ban from bannedUsers.json and writes to bannedUsers.json
								misc.BannedUsersSlice = append(misc.BannedUsersSlice[:i], misc.BannedUsersSlice[i+1:]...)
								misc.BannedUsersWrite(misc.BannedUsersSlice)
								break
							}
						}

						if banFlag == false {

							//Sends a message to the channel with error message
							_, err = s.ChannelMessageSend(m.ChannelID, mem.Username+" is not banned.")
							if err != nil {

								fmt.Println("Error: ", err)
							}
						}

					} else {

						//Sends a message to the channel with error message
						_, err = s.ChannelMessageSend(m.ChannelID, mem.Username+" is not banned.")
						if err != nil {

							fmt.Println("Error: ", err)
						}
					}

					if banFlag == true {

						//Removes the ban
						s.GuildBanDelete(config.ServerID, mem.ID)

						//Sends a message to the channel with error message
						_, err = s.ChannelMessageSend(m.ChannelID, mem.Username+"#"+mem.Discriminator+" has been unbanned.")
						if err != nil {

							fmt.Println("Error: ", err)
						}

						//Saves time of unban command usage
						t := time.Now()

						misc.MapMutex.Lock()

						//Updates unban date in memberInfo.json
						misc.MemberInfoMap[mem.ID].UnbanDate = t.Format("2006-01-02 15:04:05")

						misc.MapMutex.Unlock()

						//Writes to memberInfo.json
						misc.MemberInfoWrite(misc.MemberInfoMap)
					}
				}
			}
		}
	}
}
