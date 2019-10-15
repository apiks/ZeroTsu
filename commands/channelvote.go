package commands

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

var inChanCreation bool

// Starts a 30 hour vote in a channel with parameters that automatically creates a spoiler channel if the vote passes
func startVoteCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		peopleNum        = 7
		descriptionSlice []string
		voteChannel      channel
		controlNumVote   = 0
		controlNumUser   = 0
		controlNum       = 0
		typeFlag         bool
		admin            bool
	)

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	if !guildSettings.VoteModule {
		functionality.MapMutex.Unlock()
		return
	}

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	// Checks if the message author is an admin or not and saves it, to save operations down the line
	if functionality.HasElevatedPermissions(s, m.Author.ID, m.GuildID) {
		admin = true
	}
	functionality.MapMutex.Unlock()

	if admin {
		if len(commandStrings) == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"startvote OPTIONAL[votes required] [name] OPTIONAL[type] OPTIONAL[categoryID] + OPTIONAL[description]`\n\n"+
				"Votes required is how many thumbs up to require to create a channel. Default is 7.\nTypes are temp (deleted after 3 hours of inactivity), optin, airing and general. They are optional and default is optin. Do _not_ use types in the channel name\n"+
				"CategoryID is the ID of the category the channel will be created in. it is optional.\nDescription is the description of that channel. It is optional but _needs_ a categoryID or Type before it or it will break.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}

		command := strings.Replace(messageLowercase, guildSettings.Prefix+"startvote ", "", -1)
		commandStrings = strings.Split(command, " ")

		// Checks if [category] and [type] exist and assigns them if they do and removes them from slice
		for index := range commandStrings {
			_, err := strconv.Atoi(commandStrings[index])
			if len(commandStrings[index]) >= 17 && err == nil {
				voteChannel.Category = commandStrings[index]
				command = strings.Replace(command, commandStrings[index], "", -1)
			}

			if commandStrings[index] == "airing" ||
				commandStrings[index] == "general" ||
				commandStrings[index] == "opt-in" ||
				commandStrings[index] == "optin" ||
				commandStrings[index] == "temp" &&
					!typeFlag {

				voteChannel.Type = commandStrings[index]
				command = strings.Replace(command, commandStrings[index], "", 1)
				typeFlag = true
			}
		}

		// If either [description] or [type] exist then checks if a description is also present
		if voteChannel.Type != "" || voteChannel.Category != "" {
			if voteChannel.Category != "" {
				descriptionSlice = strings.SplitAfter(m.Content, voteChannel.Category)
			} else {
				descriptionSlice = strings.SplitAfter(m.Content, voteChannel.Type)
			}

			// Makes the description the second element of the slice above
			voteChannel.Description = descriptionSlice[1]
			// Makes a copy of description that it puts to lowercase
			descriptionLowercase := strings.ToLower(voteChannel.Description)
			// Removes description from command variable
			command = strings.Replace(command, descriptionLowercase, "", -1)
		}

		// If the first parameter is a number set that to be the votes required
		if num, err := strconv.Atoi(commandStrings[0]); err == nil {
			peopleNum = num
			// Removes the num from the command string
			command = strings.Replace(command, commandStrings[0]+" ", "", 1)
		}

		// Assigns channel name and checks if it's empty
		voteChannel.Name = command
		blank := strings.TrimSpace(voteChannel.Name) == ""

		// If the name doesn't exist, print an error
		if blank {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Channel name not parsed properly. Please use `"+guildSettings.Prefix+"startvote "+
				"OPTIONAL[votes required] [name] OPTIONAL[type] OPTIONAL[categoryID] + OPTIONAL[description]`")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	} else {
		if len(commandStrings) == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"startvote [name]`")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}

		voteChannel.Name = strings.Replace(messageLowercase, guildSettings.Prefix+"startvote ", "", -1)
		if guildSettings.VoteChannelCategory != nil {
			if guildSettings.VoteChannelCategory.ID != "" {
				voteChannel.Category = guildSettings.VoteChannelCategory.ID
			}
		}
		voteChannel.Type = "temp"
		voteChannel.Description = fmt.Sprintf("Temporary channel for %v. Will be deleted 3 hours after no message has been sent.", voteChannel.Name)

		// Required minimum number of people for a vote to pass
		peopleNum = 3
	}

	// Checks if a channel is in the process of being created before moving on (prevention against spam)
	if inChanCreation {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: A channel is in the process of being created. Please try again in 10 seconds.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls up all current server channels and checks if it exists in UserTempCha.json. If not it deletes it from storage
	cha, err := s.GuildChannels(m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Checks ongoing non-mod temp channels
	if !admin {
		functionality.MapMutex.Lock()
		for k, v := range functionality.GuildMap[m.GuildID].TempChaMap {
			exists := false
			for i := 0; i < len(cha); i++ {
				if cha[i].Name == v.RoleName {
					exists = true
					break
				}
			}
			if !exists {
				delete(functionality.GuildMap[m.GuildID].TempChaMap, k)
				_ = functionality.TempChaWrite(functionality.GuildMap[m.GuildID].TempChaMap, m.GuildID)
			}
			if !v.Elevated {
				controlNumUser++
			}
		}
		functionality.MapMutex.Unlock()

		// Prints error if the user temp channel cap (3) has been reached
		if controlNumUser > 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Maximum number of user made temp channels(3) has been reached."+
				" Please contact a mod for a new temp channel or wait for the other three to run out.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}

		// Checks if there are any current temp channel votes made by users and already created channels that have reached the cap and prints error if there are
		functionality.MapMutex.Lock()
		for _, v := range functionality.GuildMap[m.GuildID].VoteInfoMap {
			if !functionality.HasElevatedPermissions(s, v.User.ID, m.GuildID) {
				controlNumVote++
			}
		}
		functionality.MapMutex.Unlock()
		controlNum = controlNumVote + controlNumUser
		if controlNum > 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are already ongoing user temp votes that breach the cap(3) together with already created temp channels. "+
				"Please contact a mod for a new temp channel or wait for the votes or temp channels to run out.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}

	// Fixes role name bug
	role := strings.Replace(strings.TrimSpace(voteChannel.Name), " ", "-", -1)
	role = strings.Replace(role, "--", "-", -1)
	voteChannel.Name = role

	peopleNum = peopleNum + 1
	peopleNumStr := strconv.Itoa(peopleNum)
	messageReact, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v thumbs up reacts on this message will create `%v`. Time limit is 30 hours.", peopleNumStr, voteChannel.Name))
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	if messageReact == nil {
		return
	}

	_ = s.MessageReactionAdd(messageReact.ChannelID, messageReact.ID, "👍")
	messageReact, err = s.ChannelMessage(messageReact.ChannelID, messageReact.ID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	t := time.Now()

	var temp functionality.VoteInfo

	// Saves the date of removal in separate variable and then adds 30 hours to it
	thirtyHours := time.Hour * 30
	dateRemoval := t.Add(thirtyHours)

	// Assigns values to VoteInfoMap so it can be written to storage
	temp.Date = dateRemoval
	temp.Channel = voteChannel.Name
	temp.Description = voteChannel.Description
	temp.Category = voteChannel.Category
	temp.ChannelType = voteChannel.Type
	temp.VotesReq = peopleNum
	temp.MessageReact = messageReact
	temp.User = m.Author

	functionality.MapMutex.Lock()
	functionality.GuildMap[m.GuildID].VoteInfoMap[m.ID] = &temp

	// Writes to storage
	err = functionality.VoteInfoWrite(functionality.GuildMap[m.GuildID].VoteInfoMap, m.GuildID)
	if err != nil {
		functionality.MapMutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	functionality.MapMutex.Unlock()

	if !admin {
		if guildSettings.BotLog == nil {
			return
		}
		if guildSettings.BotLog.ID == "" {
			return
		}
		_, err = s.ChannelMessageSend(guildSettings.BotLog.ID, fmt.Sprintf("Vote for temp channel `%v` has been started by user %v#%v in %v.", temp.Channel, temp.User.Username, temp.User.Discriminator, functionality.ChMentionID(m.ChannelID)))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
}

// Checks if the message has enough reacts every 30 seconds, and stops if it's over the time limit
func ChannelVoteTimer(s *discordgo.Session, e *discordgo.Ready) {

	var (
		timestamp time.Time
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in ChannelVoteTimer")
		}
	}()

	for range time.NewTicker(30 * time.Second).C {
		functionality.MapMutex.Lock()
		for _, guild := range e.Guilds {

			functionality.HandleNewGuild(s, guild.ID)
			guildSettings := functionality.GuildMap[guild.ID].GetGuildSettings()
			if !guildSettings.VoteModule {
				continue
			}

			for k := range functionality.GuildMap[guild.ID].VoteInfoMap {

				t := time.Now()

				// Updates message
				messageReact, err := s.ChannelMessage(functionality.GuildMap[guild.ID].VoteInfoMap[k].MessageReact.ChannelID, functionality.GuildMap[guild.ID].VoteInfoMap[k].MessageReact.ID)
				if err != nil {
					//If message doesn't exist (was deleted or otherwise not found)
					// Deletes the vote from memory
					delete(functionality.GuildMap[guild.ID].VoteInfoMap, k)
					// Writes to storage
					_ = functionality.VoteInfoWrite(functionality.GuildMap[guild.ID].VoteInfoMap, guild.ID)
					continue
				}

				// Calculates if it's time to remove
				difference := t.Sub(functionality.GuildMap[guild.ID].VoteInfoMap[k].Date)
				if difference > 0 {
					if messageReact == nil {
						continue
					}

					// Fixes role name bugs
					role := strings.Replace(strings.TrimSpace(functionality.GuildMap[guild.ID].VoteInfoMap[k].Channel), " ", "-", -1)
					role = strings.Replace(role, "--", "-", -1)

					numStr := strconv.Itoa(functionality.GuildMap[guild.ID].VoteInfoMap[k].VotesReq - 1)
					_, err = s.ChannelMessageSend(messageReact.ChannelID, "Channel vote has ended. `"+functionality.GuildMap[guild.ID].VoteInfoMap[k].Channel+"` has failed to "+
						"gather the necessary "+numStr+" votes.")
					if err != nil {
						functionality.LogError(s, guildSettings.BotLog, err)
						continue
					}

					// Deletes the vote from memory
					delete(functionality.GuildMap[guild.ID].VoteInfoMap, k)

					// Writes to storage
					_ = functionality.VoteInfoWrite(functionality.GuildMap[guild.ID].VoteInfoMap, guild.ID)
					continue
				}

				if messageReact == nil {
					continue
				}
				if messageReact.Reactions == nil {
					continue
				}
				// Checks if the vote was successful and executes if so
				if messageReact.Reactions[0].Count < functionality.GuildMap[guild.ID].VoteInfoMap[k].VotesReq {
					continue
				}

				// Tell the startvote command to wait for this command to finish before moving on
				inChanCreation = true

				var (
					message discordgo.Message
					author  discordgo.User
				)

				// Fixes role name bugs
				role := strings.Replace(strings.TrimSpace(functionality.GuildMap[guild.ID].VoteInfoMap[k].Channel), " ", "-", -1)
				role = strings.Replace(role, "--", "-", -1)

				// Removes all hyphen prefixes and suffixes because discord cannot handle them
				for strings.HasPrefix(role, "-") || strings.HasSuffix(role, "-") {
					role = strings.TrimPrefix(role, "-")
					role = strings.TrimSuffix(role, "-")
				}

				functionality.GuildMap[guild.ID].VoteInfoMap[k].Channel = role

				// Allows entry to be deleted from memory now, rather than later, avoiding potential bugs if the below commands don't work
				temp := functionality.GuildMap[guild.ID].VoteInfoMap[k]
				delete(functionality.GuildMap[guild.ID].VoteInfoMap, k)
				_ = functionality.VoteInfoWrite(functionality.GuildMap[guild.ID].VoteInfoMap, guild.ID)

				// Create command
				author.ID = s.State.User.ID
				message.ChannelID = temp.MessageReact.ChannelID
				message.GuildID = guild.ID
				message.Author = &author
				if temp.Category == "" {
					message.Content = fmt.Sprintf("%vcreate %v %v %v", guildSettings.Prefix, temp.Channel, temp.ChannelType, temp.Description)
				} else {
					message.Content = fmt.Sprintf("%vcreate %v %v %v %v", guildSettings.Prefix, temp.Channel, temp.ChannelType, temp.Category, temp.Description)
				}

				createChannelCommand(s, &message)

				time.Sleep(500 * time.Millisecond)

				// Sortroles command if optin, airing or temp
				if temp.ChannelType != "general" {
					message.Content = guildSettings.Prefix + "sortroles"
					sortRolesCommand(s, &message)
				}

				time.Sleep(500 * time.Millisecond)

				// Sortcategory command if category exists and it's not temp
				if temp.Category != "" && temp.ChannelType != "temp" && temp.ChannelType != "temporary" {
					message.Content = guildSettings.Prefix + "sortcategory " + temp.Category
					sortCategoryCommand(s, &message)
				}

				if temp.ChannelType == "general" {
					_, err = s.ChannelMessageSend(messageReact.ChannelID, "Channel `"+temp.Channel+"` was successfully created!")
					if err != nil {
						functionality.LogError(s, guildSettings.BotLog, err)
						inChanCreation = false
						continue
					}
				} else {
					_, err = s.ChannelMessageSend(messageReact.ChannelID, "Channel `"+temp.Channel+"` was successfully created! Those that have voted were given the role. Use `"+
						guildSettings.Prefix+"join "+role+"` if you do not have it.")
					if err != nil {
						functionality.LogError(s, guildSettings.BotLog, err)
						inChanCreation = false
						continue
					}
				}

				time.Sleep(200 * time.Millisecond)

				if temp.ChannelType != "general" {
					roles, err := s.GuildRoles(guild.ID)
					if err != nil {
						functionality.LogError(s, guildSettings.BotLog, err)
						inChanCreation = false
						continue
					}
					for i := 0; i < len(roles); i++ {
						if roles[i].Name == role {
							role = roles[i].ID
							break
						}
					}

					// Gets the users who voted and gives them the role
					users, err := s.MessageReactions(temp.MessageReact.ChannelID, temp.MessageReact.ID, "👍", 100)
					if err != nil {
						functionality.LogError(s, guildSettings.BotLog, err)
						inChanCreation = false
						continue
					}

					for i := 0; i < len(users); i++ {
						if users[i].ID == s.State.User.ID {
							continue
						}

						err := s.GuildMemberRoleAdd(guild.ID, users[i].ID, role)
						if err != nil {
							functionality.LogError(s, guildSettings.BotLog, err)
							continue
						}
					}
				}

				inChanCreation = false

				if temp.ChannelType == "temp" || temp.ChannelType == "temporary" {
					if guildSettings.BotLog == nil {
						continue
					}
					if guildSettings.BotLog.ID != "" {
						continue
					}
					_, err = s.ChannelMessageSend(guildSettings.BotLog.ID, fmt.Sprintf("Temp channel `%v` has been created from a vote by user %v#%v.", temp.Channel, temp.User.Username, temp.User.Discriminator))
					if err != nil {
						functionality.LogError(s, guildSettings.BotLog, err)
						continue
					}
					continue
				}
			}

			cha, err := s.GuildChannels(guild.ID)
			if err != nil {
				continue
			}

			for k, v := range functionality.GuildMap[guild.ID].TempChaMap {
				for i := 0; i < len(cha); i++ {
					if cha[i].Name == v.RoleName {
						mess, err := s.ChannelMessages(cha[i].ID, 1, "", "", "")
						if err != nil {
							functionality.LogError(s, guildSettings.BotLog, err)
							continue
						}

						// Fetches the properly parsed timestamp from discord, else uses channel creation
						if len(mess) != 0 {
							timestamp, err = mess[0].Timestamp.Parse()
							if err != nil {
								functionality.LogError(s, guildSettings.BotLog, err)
								continue
							}
						} else {
							timestamp = v.CreationDate
						}

						// Adds how long before last message for channel to be deletes
						timestamp = timestamp.Add(time.Hour * 3)
						t := time.Now()

						// Calculates if it's time to remove
						difference := t.Sub(timestamp)
						if difference > 0 {

							if guildSettings.BotLog != nil {
								if guildSettings.BotLog.ID != "" {
									_, err = s.ChannelMessageSend(guildSettings.BotLog.ID, fmt.Sprintf("Temp channel `%v` has been deleted due to being inactive for 3 hours.", cha[i].Name))
									if err != nil {
										functionality.LogError(s, guildSettings.BotLog, err)
										continue
									}
								}
							}

							// Deletes channel and role
							_, err := s.ChannelDelete(cha[i].ID)
							if err != nil {
								functionality.LogError(s, guildSettings.BotLog, err)
								continue
							}

							err = s.GuildRoleDelete(guild.ID, k)
							if err != nil {
								functionality.LogError(s, guildSettings.BotLog, err)
								continue
							}

							delete(functionality.GuildMap[guild.ID].TempChaMap, k)
							_ = functionality.TempChaWrite(functionality.GuildMap[guild.ID].TempChaMap, guild.ID)
						}
					}
				}
			}
		}
		functionality.MapMutex.Unlock()
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute: startVoteCommand,
		Trigger: "startvote",
		Desc:    "Starts a vote for channel creation [VOTE]",
		Module:  "channel",
	})
}
