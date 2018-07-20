package commands

import (
	"strings"
	"fmt"
	"math"

	"github.com/bwmarrin/discordgo"

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

					// Sets whois message
					message := "**User:** " + mem.Mention() + "\n\n**Past Usernames:** " + pastUsernames +
						"\n\n**Past Nicknames:** " + pastNicknames + "\n\n**Warnings:** " + warnings +
						"\n\n**Kicks:** " + kicks + "\n\n**Bans:** " + bans +
						"\n\n**Join Date:** " + user.JoinDate + "\n\n**Verification Date:** " +
						user.VerifiedDate + "\n\n**Reddit Account:** " +
						"<https://reddit.com/u/" + user.RedditUsername + ">"

					// Alt check
					alts := CheckAltAccountWhois(mem.ID)

					// If there's more than one account with that reddit username print a message
					if len(alts) > 1 {

						// Forms the success string
						success := "\n\n**Alts:** \n\n"
						for i := 0; i < len(alts); i++ {

							success = success + "<@" + alts[i] + "> \n"
						}

						// Adds the alts to the whois message
						message = message + success

						// Resets alts variable
						alts = nil
					}

					var splitMessage []string

					// Splits the message if it's over 1950 characters
					if len(message) > 195 {

						splitMessage = SplitLongMessage(message)
					}

					if splitMessage != nil {
						for i := 0; i < len(splitMessage); i++ {

							_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
							if err != nil {

								_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot send whois message.")
								if  err != nil {

									fmt.Println(err)
								}
							}
						}
					} else {

						_, err := s.ChannelMessageSend(m.ChannelID, message)
						if err != nil {

							_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot send whois message.")
							if err != nil {

								fmt.Println(err)
							}
						}
					}
				}
			}
		}
	}
}

// SplitLongMessage takes a message and splits it if it's longer than 1950. By Kagumi
func SplitLongMessage(message string) (split []string) {
	const maxLength = 1950
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

// Function that iterates through memberInfo.json and checks for any alt accounts for that ID. Whois version
func CheckAltAccountWhois(id string) []string {

	// Initializes alts string slice to hold IDs of alts of that reddit username
	var alts []string

	// Reads memberInfo
	misc.MemberInfoRead()

	// Iterates through all users in memberInfo.json
	for userOne := range misc.MemberInfoMap {

		// Checks if the current user has the same reddit username as id string user
		if misc.MemberInfoMap[userOne].RedditUsername == misc.MemberInfoMap[id].RedditUsername {

			alts = append(alts, misc.MemberInfoMap[userOne].ID)
		}
	}

	if len(alts) > 1 {

		return alts
	} else {

		return nil
	}
}
