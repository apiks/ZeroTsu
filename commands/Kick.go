package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

//Kick Command with reason
func KickHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

			//Checks if BotPrefix + kick was used
			if strings.HasPrefix(messageLowercase, config.BotPrefix+"kick ") && (messageLowercase != (config.BotPrefix + "kick")) {

				if m.Author.ID == config.BotID {
					return
				}

				var (
					user   string
					reason string
				)

				//Pulls the user and reason from strings after "kick ". Gives error if not enough parameters
				userSlice := strings.SplitN(m.Content, " ", 3)
				if len(userSlice) == 3 {

					user = userSlice[1]
					reason = userSlice[2]
				} else {

					//Sends a message to the channel with error message
					_, err = s.ChannelMessageSend(m.ChannelID, "Error. Please use `"+config.BotPrefix+"kick [@user or userID] [reason]` format.")
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
						kicks   string
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
							fmt.Print("User not found in memberInfo. Initializing user.")
							_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo. Initializing user.")

							misc.InitializeUser(userMem)

							//Assigns to user variable
							user = misc.MemberInfoMap[mem.ID]

							//If user is not in the server it cannot initialize
						} else {

							_, err = s.ChannelMessageSend(m.ChannelID, "Error: User not found in memberInfo. Cannot kick until joins server.")
							return
						}

						//Writes to memberInfo.json
						misc.MemberInfoWrite(misc.MemberInfoMap)
					}

					misc.MapMutex.Lock()

					//Adds kicks to user memberInfo
					user.Kicks = append(user.Kicks, reason)

					misc.MapMutex.Unlock()

					//Pulls the guild Name
					guildName, err := s.Guild(config.ServerID)
					if err != nil {

						fmt.Println("Error: ", err)
					}
					name := guildName.Name

					success = "You have been kicked from " + name + ":\n**" + reason + "**"

					//Creates a DM connection and assigns it to dm
					dm, err := s.UserChannelCreate(mem.ID)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Sends a message to that DM connection for kick
					_, err = s.ChannelMessageSend(dm.ID, success)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Sends embed bot-log message
					KickEmbed(s, m, mem, reason)

					//Kicks user
					err = s.GuildMemberDeleteWithReason(config.ServerID, mem.ID, reason)
					if err != nil {

						fmt.Println("Error kicking: ", err)
					}

					misc.MapMutex.Lock()

					//Puts kicks into a string
					if len(user.Kicks) != 0 {
						for i := 0; i < len(user.Kicks); i++ {

							if len(kicks) == 0 {

								kicks = user.Kicks[i]
							} else {

								kicks = kicks + ", " + user.Kicks[i]
							}
						}

						kicks = kicks + ", " + reason
					} else {

						kicks = reason
					}

					misc.MapMutex.Unlock()

					//Writes memberInfo.json
					misc.MemberInfoWrite(misc.MemberInfoMap)

					if err != nil {

						fmt.Println("Error: ", err)
					}
				}
			}
		}
	}
}

func KickEmbed(s *discordgo.Session, m *discordgo.MessageCreate, mem *discordgo.User, reason string) {

	//Initializing needed variables for the embed
	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail

		//Embed slice and its fields
		embedField       []*discordgo.MessageEmbedField
		embedFieldUserID discordgo.MessageEmbedField
		embedFieldReason discordgo.MessageEmbedField
	)

	//Saves user avatar as thumbnail
	embedThumbnail.URL = mem.AvatarURL("128")

	//Sets field titles
	embedFieldUserID.Name = "User ID:"
	embedFieldReason.Name = "Reason:"

	//Sets field content
	embedFieldUserID.Value = mem.ID
	embedFieldReason.Value = reason

	//Sets field inline
	embedFieldUserID.Inline = true
	embedFieldReason.Inline = true

	//Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldUserID)
	embedField = append(embedField, &embedFieldReason)

	//Sets embed title and its description (which it uses the same way as a field)
	embedMess.Title = mem.Username + "#" + mem.Discriminator + " was kicked by " + m.Author.Username

	//Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	//Send embed in bot-log channel
	_, err := s.ChannelMessageSendEmbed(config.BotLogID, &embedMess)
	if err != nil {

		fmt.Println("Error: ", err)
	}
}
