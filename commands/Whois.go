package commands

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"math"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

//Sends user information as embed message
func WhoisHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

			//Checks if BotPrefix + whois was used
			if strings.HasPrefix(messageLowercase, config.BotPrefix+"whois ") && (messageLowercase != (config.BotPrefix + "whois")) {

				if m.Author.ID == config.BotID {
					return
				}

				// Saves program from panic and continues running normally without executing the command if it happens
				defer func() {
					if r := recover(); r != nil {

						fmt.Println(r)
					}
				}()

				//Pulls the user from strings after "whois "
				user := strings.Replace(messageLowercase, config.BotPrefix+"whois ", "", 1)

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

					//Initializes message needed variables
					var (
						pastUsernames string
						pastNicknames string
						warnings      string
						kicks         string
						bans          string
						unbanDate     string
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

							//Initializes user if he doesn't exist and is in server and stops command
							fmt.Print("User: " + userMem.User.Username + " not found in memberInfo. Initializing user.")
							_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo. Initializing user.")

							misc.InitializeUser(userMem)
							misc.MemberInfoWrite(misc.MemberInfoMap)

							return
							//If user is not in the server it cannot initialize
						} else {

							_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo. Cannot whois until user joins server.")
							return
						}

						//Writes to memberInfo.json
						misc.MemberInfoWrite(misc.MemberInfoMap)
					}

					misc.MapMutex.Lock()

					//Puts past usernames into a string
					if len(user.PastUsernames) != 0 {
						for i := 0; i < len(user.PastUsernames); i++ {

							if len(pastUsernames) == 0 {

								pastUsernames = user.PastUsernames[i]
							} else {

								pastUsernames = pastUsernames + ", " + user.PastUsernames[i]
							}
						}
					} else {

						pastUsernames = "None"
					}

					//Puts past nicknames into a string
					if len(user.PastNicknames) != 0 {
						for i := 0; i < len(user.PastNicknames); i++ {

							if len(pastNicknames) == 0 {

								pastNicknames = user.PastNicknames[i]
							} else {

								pastNicknames = pastNicknames + ", " + user.PastNicknames[i]
							}
						}
					} else {

						pastNicknames = "None"
					}

					//Puts warnings into a slice
					if len(user.Warnings) != 0 {
						for i := 0; i < len(user.Warnings); i++ {

							if len(warnings) == 0 {

								warnings = user.Warnings[i]
							} else {

								warnings = warnings + ", " + user.Warnings[i]
							}
						}
					} else {

						warnings = "None"
					}

					//Puts kicks into a slice
					if len(user.Kicks) != 0 {
						for i := 0; i < len(user.Kicks); i++ {

							if len(kicks) == 0 {

								kicks = user.Kicks[i]
							} else {

								kicks = kicks + ", " + user.Kicks[i]
							}
						}
					} else {

						kicks = "None"
					}

					//Puts bans into a slice
					if len(user.Bans) != 0 {
						for i := 0; i < len(user.Bans); i++ {

							if len(bans) == 0 {

								bans = user.Bans[i]
							} else {

								bans = bans + ", " + user.Bans[i]
							}
						}
					} else {

						bans = "None"
					}

					//Puts unban Date into a separate string variable
					unbanDate = user.UnbanDate
					if unbanDate == "" {

						unbanDate = "User has never been banned."
					}

					misc.MapMutex.Unlock()

					var splitUsernames []string
					var splitNicknames []string

					// Splits past usernames if they're over 1800 characters to avoid message limit
					if len(pastUsernames) > 1800 {

						splitUsernames = SplitLongMessage(pastUsernames)
					}

					// Splits past nicknames if they're over 1800 characters to avoid message limit
					if len(pastNicknames) > 1800 {

						splitNicknames = SplitLongMessage(pastNicknames)
					}

					// Iterates through all split usernames and sends message and sends message to chat for whois command
					if splitUsernames != nil {
						for i := 0; i < len(splitUsernames); i++ {

							if i == 0 {
								// Prints the user information in simple text
								_, err := s.ChannelMessageSend(m.ChannelID, "User: "+mem.Mention()+"\n\n *Past Usernames:* "+splitUsernames[0])
								if err != nil {

									_, err = s.ChannelMessageSend(m.ChannelID, "Error: Cannot whois user. Please check the first print function in the code.")
								}
							} else {

								// Prints the user information in simple text
								_, err := s.ChannelMessageSend(m.ChannelID, "\n" + splitUsernames[i])
								if err != nil {

									_, err = s.ChannelMessageSend(m.ChannelID, "Error: Cannot whois user. Please check the first print function in the code.")
								}
							}
						}
					} else {
						// Prints the user information in simple text
						_, err := s.ChannelMessageSend(m.ChannelID, "User: "+mem.Mention()+"\n\n *Past Usernames:* "+pastUsernames)
						if err != nil {

							_, err = s.ChannelMessageSend(m.ChannelID, "Error: Cannot whois user. Please check the first print function in the code.")
						}
					}

					// Iterates through all split nicknames and sends message to chat for whois command
					if splitNicknames != nil {
						for i := 0; i < len(splitNicknames); i++ {

							if i == 0 {

								// Prints the user information in simple text
								_, err = s.ChannelMessageSend(m.ChannelID, "\n\n*Past Nicknames:* " + splitNicknames[0])
								if err != nil {

									_, err = s.ChannelMessageSend(m.ChannelID, "Error: Cannot whois user. Please check the second print function in the code."+err.Error())
								}
							} else {

								// Prints the user information in simple text
								_, err = s.ChannelMessageSend(m.ChannelID, "\n" + splitNicknames[i])
								if err != nil {

									_, err = s.ChannelMessageSend(m.ChannelID, "Error: Cannot whois user. Please check the second print function in the code."+err.Error())
								}
							}
						}
					} else {

						// Prints the user information in simple text
						_, err = s.ChannelMessageSend(m.ChannelID, "*Past Nicknames:* " + pastNicknames+
							"\n\n *Join Date:* "+ user.JoinDate + "\n *Reddit Account:* " + user.RedditUsername)
						if err != nil {

							_, err = s.ChannelMessageSend(m.ChannelID, "Error: Cannot whois user. Please check the second print function in the code."+err.Error())
						}
					}
				}
			}
		}
	}
}

// SplitLongMessage takes a message and splits it if it's longer than 1850. By Kagumi
func SplitLongMessage(message string) (split []string) {
	const maxLength = 1800
	if len(message) > maxLength {
		partitions := len(message) / maxLength
		if math.Mod(float64(len(message)), maxLength) > 0 {
			partitions++
		}
		split = make([]string, partitions)
		for i := 0; i < partitions; i++ {
			if i == partitions-1 {
				split[i] = message[i*maxLength:]
				break
			}
			split[i] = message[i*maxLength : (i+1)*maxLength]
		}
	} else {
		split = make([]string, 1)
		split[0] = message
	}
	return
}