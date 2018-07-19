package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

//Ban Command with reason
func BanHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

			//Checks if BotPrefix + ban was used
			if strings.HasPrefix(messageLowercase, config.BotPrefix+"ban ") && (messageLowercase != (config.BotPrefix + "ban")) {

				if m.Author.ID == config.BotID {
					return
				}

				var (
					user   string
					reason string
					length string

					error bool
				)

				//Pulls the user time and reason from messageLowercase
				userSlice := strings.SplitN(messageLowercase, " ", 4)

				//Checks if it has all parameters, else error
				if len(userSlice) == 4 {

					user = userSlice[1]
					length = userSlice[2]
					reason = userSlice[3]

				} else {

					error = true

					//Sends a message to the channel with error message
					_, err = s.ChannelMessageSend(m.ChannelID, "Error. Please use `"+config.BotPrefix+"ban [@user or userID] [time] [reason]` format. \n\n"+
						"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes. Use 0d for permanent.")
					if err != nil {

						fmt.Println("Error: ", err)
					}
				}

				//Pulls the reason in uppercase
				userSlice = strings.SplitN(m.Content, " ", 4)

				//Checks if it has all parameters, else error
				if len(userSlice) == 4 {

					//Assigns the reason in uppercase
					reason = userSlice[3]

				} else if error == false {

					//Sends a message to the channel with error message
					_, err = s.ChannelMessageSend(m.ChannelID, "Error. Please use `"+config.BotPrefix+"ban [@user or userID] [time] [reason]` format. \n\n"+
						"Time is in #w#d#h#m format, such as 2w1d12h30m for 2 weeks, 1 day, 12 hours, 30 minutes. Use 0d for permanent.")
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
				if err != nil {

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
						bans    string
						temp    misc.BannedUsers
						success string
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
							fmt.Print("User: not found in memberInfo. Initializing user.")
							_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo. Initializing user.")

							misc.InitializeUser(userMem)

							//Assigns to user variable
							user = misc.MemberInfoMap[mem.ID]

							//If user is not in the server it cannot initialize
						} else {

							_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo. Cannot ban until joins server.")
							return
						}

						//Writes to memberInfo.json
						misc.MemberInfoWrite(misc.MemberInfoMap)
					}

					misc.MapMutex.Lock()

					//Adds unban date to memberInfo and checks if perma
					user.Bans = append(user.Bans, reason)
					UnbanDate, perma := misc.ResolveTimeFromString(length)

					if perma == false {
						user.UnbanDate = UnbanDate.Format("2006-01-02 15:04:05")

					} else {

						user.UnbanDate = "_Never_"
					}

					misc.MapMutex.Unlock()

					//Saves the details in temp
					temp.ID = user.ID
					temp.User = user.Username

					if perma == false {

						temp.UnbanDate = UnbanDate
					} else {

						temp.UnbanDate = time.Date(9999, 9, 9, 9, 9, 9, 9, time.Local)
					}

					//Reads bannedUsers.json
					misc.BannedUsersRead()

					//Adds the now banned user to bannedUsers.json
					misc.BannedUsersSlice = append(misc.BannedUsersSlice, temp)

					//Writes the new bannedUsers.json to file
					misc.BannedUsersWrite(misc.BannedUsersSlice)

					//Pulls the guild Name
					guildName, err := s.Guild(config.ServerID)
					if err != nil {

						fmt.Println("Error: ", err)
					}
					name := guildName.Name

					//Assigns success print string for user
					if perma == false {

						success = "You have been banned from " + name + ": **" + reason + "**\n\nUntil: _" + UnbanDate.Format("2006-01-02 15:04:05") + "_"
					} else {

						success = "You have been banned from " + name + ": **" + reason + "**\n\nUntil: _Forever_ \n\nIf you would like to appeal, use modmail at <https://reddit.com/r/anime>"
					}

					//Creates a DM connection and assigns it to dm
					dm, err := s.UserChannelCreate(user.ID)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Sends a message to that DM connection for ban
					_, err = s.ChannelMessageSend(dm.ID, success)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Sends embed bot-log message
					BanEmbed(s, m, mem, reason, length)

					//Bans user
					err = s.GuildBanCreateWithReason(config.ServerID, mem.ID, reason, 0)
					if err != nil {

						fmt.Println("Error banning: ", err)
					}

					misc.MapMutex.Lock()

					//Puts bans into a string
					if len(user.Bans) != 0 {
						for i := 0; i < len(user.Bans); i++ {

							if len(bans) == 0 {

								bans = user.Bans[i]
							} else {

								bans = bans + ", " + user.Bans[i]
							}
						}

						bans = bans + ", " + reason
					} else {

						bans = reason
					}

					misc.MapMutex.Unlock()

					//Writes memberInfo.json
					misc.MemberInfoWrite(misc.MemberInfoMap)

					//Sends a message to bot-log for ban
					if perma == false {

						_, err = s.ChannelMessageSend(m.ChannelID, user.Username+"#"+user.Discrim+" has been banned by "+
							m.Author.Username+" until _"+UnbanDate.Format("2006-01-02 15:04:05")+"_")
					} else {

						_, err = s.ChannelMessageSend(m.ChannelID, user.Username+"#"+user.Discrim+" has been permabanned by "+
							m.Author.Username)
					}

					if err != nil {

						fmt.Println("Error: ", err)
					}
				}
			}
		}
	}
}

func BanEmbed(s *discordgo.Session, m *discordgo.MessageCreate, mem *discordgo.User, reason string, length string) {

	//Initializing needed variables for the embed
	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail

		//Embed slice and its fields
		embedField         []*discordgo.MessageEmbedField
		embedFieldUserID   discordgo.MessageEmbedField
		embedFieldReason   discordgo.MessageEmbedField
		embedFieldDuration discordgo.MessageEmbedField
	)

	//Saves user avatar as thumbnail
	embedThumbnail.URL = mem.AvatarURL("128")

	//Sets field titles
	embedFieldUserID.Name = "User ID:"
	embedFieldReason.Name = "Reason:"
	embedFieldDuration.Name = "Duration:"

	//Sets field content
	embedFieldUserID.Value = mem.ID
	embedFieldReason.Value = reason
	embedFieldDuration.Value = length

	//Sets field inline
	embedFieldUserID.Inline = true
	embedFieldReason.Inline = true
	embedFieldDuration.Inline = true

	//Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldUserID)
	embedField = append(embedField, &embedFieldDuration)
	embedField = append(embedField, &embedFieldReason)

	//Sets embed title and its description (which it uses the same way as a field)
	embedMess.Title = mem.Username + "#" + mem.Discriminator + " was banned by " + m.Author.Username

	//Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	//Send embed in bot-log channel
	_, err := s.ChannelMessageSendEmbed(config.BotLogID, &embedMess)
	if err != nil {

		fmt.Println("Error: ", err)
	}
}
