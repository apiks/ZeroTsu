package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

//Adds a warning to a user in memberInfo.json and tells him
func IssueWarningHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

			//Checks if BotPrefix + issuewarning was used
			if strings.HasPrefix(messageLowercase, config.BotPrefix+"issuewarning ") && (messageLowercase != (config.BotPrefix + "issuewarning")) {

				if m.Author.ID == config.BotID {
					return
				}

				var (
					user    string
					warning string
				)

				//Pulls the user and warning from strings after "issuewarning "
				userSlice := strings.SplitN(m.Content, " ", 3)

				if len(userSlice) == 3 {

					user = userSlice[1]
					warning = userSlice[2]
				} else {

					//Sends a message to the channel with error message
					_, err = s.ChannelMessageSend(m.ChannelID, "Error. Please use `"+config.BotPrefix+"issuewarning [@user or userID] [warning]` format.")
					if err != nil {

						fmt.Println("Error: ", err)
					}
				}

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
				if err != nil && user != "" {

					fmt.Println("Error: ", err)

					if user != "" {

						//Sends a message to the channel with error message
						_, err = s.ChannelMessageSend(m.ChannelID, user+" is not a valid user. Use `@user` or `user ID`.")
						if err != nil {

							fmt.Println("Error: ", err)
						}
					}
				} else if err == nil {

					var (
						warnings string
						success  string
					)

					//Reads memberInfo.json
					misc.MemberInfoRead()

					//Checks if user is in MemberInfo and assigns to user variable. Else initializes user.
					user, ok := misc.MemberInfoMap[mem.ID]
					if !ok {

						//Pulls info on user
						userMem, err := s.State.Member(config.ServerID, mem.ID)
						if err != nil {
							userMem, err = s.GuildMember(config.ServerID, mem.ID)
							if err != nil {
								fmt.Println(err.Error())
							}
						}

						if userMem != nil {

							//Initializes user if he doesn't exist and is in server
							fmt.Print("User not found in memberInfo. Initializing user.")
							_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo. Initializing user.")

							misc.InitializeUser(userMem)

							//Assigns to user variable
							user = misc.MemberInfoMap[mem.ID]

							//If user is not in the server it cannot initialize
						} else {

							_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo. Cannot warn until joins server.")
							return
						}

						//Writes to memberInfo.json
						misc.MemberInfoWrite(misc.MemberInfoMap)
					}

					misc.MapMutex.Lock()

					//Adds kicks to user memberInfo
					user.Warnings = append(user.Warnings, warning)

					misc.MapMutex.Unlock()

					//Mod success string
					success = user.Username + " has been warned with: " + "`" + warning + "`"

					//Sends a message to the message channel of success
					_, err = s.ChannelMessageSend(m.ChannelID, success)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Pulls the guild Name
					guildName, err := s.Guild(config.ServerID)
					if err != nil {

						fmt.Println("Error: ", err)
					}
					name := guildName.Name

					//User success string
					success = "You have been warned on " + name + ":\n`" +
						warning + "`"

					//Creates a DM connection and assigns it to dm
					dm, err := s.UserChannelCreate(mem.ID)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Sends a message to that DM connection for warning
					_, err = s.ChannelMessageSend(dm.ID, success)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					misc.MapMutex.Lock()

					//Puts warnings into a string
					if len(user.Warnings) != 0 {
						for i := 0; i < len(user.Warnings); i++ {

							if len(warnings) == 0 {

								warnings = user.Warnings[i]
							} else {

								warnings = warnings + ", " + user.Warnings[i]
							}
						}

						warnings = warnings + ", " + warning
					} else {

						warnings = warning
					}

					misc.MapMutex.Unlock()

					//Writes memberInfo.json
					misc.MemberInfoWrite(misc.MemberInfoMap)
				}
			}
		}
	}
}
