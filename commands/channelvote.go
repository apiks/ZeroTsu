package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"log"
	"regexp"
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

	guildSettings := db.GetGuildSettings(m.GuildID)

	// Checks if the message author is an admin or not and saves it, to save operations down the line
	if functionality.HasElevatedPermissions(s, m.Author.ID, m.GuildID) {
		admin = true
	}

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	if admin {
		if len(commandStrings) == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"startvote OPTIONAL[votes required] [name] OPTIONAL[type] OPTIONAL[categoryID] + OPTIONAL[description]`\n\n"+
				"Votes required is how many thumbs up to require to create a channel. Default is 7.\nTypes are temp (deleted after 3 hours of inactivity), optin, airing and general. They are optional and default is optin. Do _not_ use types in the channel name\n"+
				"CategoryID is the ID of the category the channel will be created in. it is optional.\nDescription is the description of that channel. It is optional but _needs_ a categoryID or Type before it or it will break.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}

		command := strings.Replace(strings.ToLower(m.Content), guildSettings.GetPrefix()+"startvote ", "", -1)
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
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Channel name not parsed properly. Please use `"+guildSettings.GetPrefix()+"startvote "+
				"OPTIONAL[votes required] [name] OPTIONAL[type] OPTIONAL[categoryID] + OPTIONAL[description]`")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	} else {
		if len(commandStrings) == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"startvote [name]`")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}

		voteChannel.Name = strings.Replace(strings.ToLower(m.Content), guildSettings.GetPrefix()+"startvote ", "", -1)
		if guildSettings.GetVoteChannelCategory() != (entities.Cha{}) {
			if guildSettings.GetVoteChannelCategory().GetID() != "" {
				voteChannel.Category = guildSettings.GetVoteChannelCategory().GetID()
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
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Pulls up all current server channels and checks if it exists in UserTempCha.json. If not it deletes it from storage
	cha, err := s.GuildChannels(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Checks ongoing non-mod temp channels
	if !admin {
		entities.Guilds.RLock()
		entities.Guilds.DB[m.GuildID].RLock()
		for k, v := range entities.Guilds.DB[m.GuildID].GetTempChaMap() {
			if v == nil {
				continue
			}

			exists := false
			for i := 0; i < len(cha); i++ {
				if cha[i].Name == v.GetRoleName() {
					exists = true
					break
				}
			}
			if !exists {
				entities.Guilds.DB[m.GuildID].RUnlock()
				entities.Guilds.RUnlock()
				err := db.SetGuildTempChannel(m.GuildID, k, v, true)
				if err != nil {
					log.Println(err)
				}
				entities.Guilds.RLock()
				entities.Guilds.DB[m.GuildID].RLock()
			}
			if !v.GetElevated() {
				controlNumUser++
			}
		}
		entities.Guilds.DB[m.GuildID].RUnlock()
		entities.Guilds.RUnlock()

		// Prints error if the user temp channel cap (3) has been reached
		if controlNumUser > 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Maximum number of user made temp channels(3) has been reached."+
				" Please contact a mod for a new temp channel or wait for the other three to run out.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}

		// Checks if there are any current temp channel votes made by users and already created channels that have reached the cap and prints error if there are

		guildVoteInfoMap := db.GetGuildVoteInfo(m.GuildID)
		for _, vote := range guildVoteInfoMap {
			if vote == nil {
				continue
			}

			if !functionality.HasElevatedPermissions(s, vote.GetUser().ID, m.GuildID) {
				controlNumVote++
			}
		}

		controlNum = controlNumVote + controlNumUser
		if controlNum > 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are already ongoing user temp votes that breach the cap(3) together with already created temp channels. "+
				"Please contact a mod for a new temp channel or wait for the votes or temp channels to run out.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}

	// Fixes role name bug
	role := strings.Replace(strings.TrimSpace(voteChannel.Name), " ", "-", -1)
	role = strings.Replace(role, "--", "-", -1)
	reg, err := regexp.Compile("[^a-zA-Z0-9-]+")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
	role = reg.ReplaceAllString(role, "")
	voteChannel.Name = role

	peopleNum = peopleNum + 1
	peopleNumStr := strconv.Itoa(peopleNum)
	messageReact, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v thumbs up reacts on this message will create `%v`. Time limit is 30 hours.", peopleNumStr, voteChannel.Name))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	if messageReact == nil {
		return
	}

	_ = s.MessageReactionAdd(messageReact.ChannelID, messageReact.ID, "👍")
	messageReact, err = s.ChannelMessage(messageReact.ChannelID, messageReact.ID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	t := time.Now()

	// Saves the date of removal in separate variable and then adds 30 hours to it
	thirtyHours := time.Hour * 30
	dateRemoval := t.Add(thirtyHours)

	err = db.SetGuildVoteInfoChannel(m.GuildID, m.ID, entities.NewVoteInfo(dateRemoval, voteChannel.Name, voteChannel.Type, voteChannel.Category, voteChannel.Description, peopleNum, messageReact, m.Author))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	if !admin {
		if guildSettings.BotLog == (entities.Cha{}) {
			return
		}
		if guildSettings.BotLog.GetID() == "" {
			return
		}
		_, err = s.ChannelMessageSend(guildSettings.BotLog.GetID(), fmt.Sprintf("Vote for temp channel `%v` has been started by user %v#%v in %v.", voteChannel.Name, m.Author.Username, m.Author.Discriminator, common.ChMentionID(m.ChannelID)))
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
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
		for _, guild := range e.Guilds {
			entities.HandleNewGuild(guild.ID)

			guildSettings := db.GetGuildSettings(guild.ID)
			if !guildSettings.GetVoteModule() {
				continue
			}
			t := time.Now()
			guildVoteInfoMap := db.GetGuildVoteInfo(guild.ID)

			for messID, vote := range guildVoteInfoMap {
				if vote == nil {
					continue
				}

				// Updates message
				messageReact, err := s.ChannelMessage(vote.GetMessageReact().ChannelID, vote.GetMessageReact().ID)
				if err != nil {
					// If message doesn't exist (was deleted or otherwise not found)
					// Deletes the vote from memory
					err = db.SetGuildVoteInfoChannel(guild.ID, messID, vote, true)
					if err != nil {
						log.Println(err)
					}
					continue
				}

				// Calculates if it's time to remove
				difference := t.Sub(vote.GetDate())
				if difference > 0 {
					if messageReact == nil {
						continue
					}

					// Fixes role name bugs
					role := strings.Replace(strings.TrimSpace(vote.GetChannel()), " ", "-", -1)
					role = strings.Replace(role, "--", "-", -1)
					reg, err := regexp.Compile("[^a-zA-Z0-9-]+")
					if err != nil {
						common.LogError(s, guildSettings.BotLog, err)
						continue
					}
					role = reg.ReplaceAllString(role, "")

					_, err = s.ChannelMessageSend(messageReact.ChannelID, fmt.Sprintf("Channel vote has ended. %s has failed to gather the necessary %d votes.", vote.GetChannel(), vote.GetVotesReq()))
					if err != nil {
						common.LogError(s, guildSettings.BotLog, err)
						continue
					}

					// Deletes vote
					err = db.SetGuildVoteInfoChannel(guild.ID, messID, vote, true)
					if err != nil {
						log.Println(err)
					}

					continue
				}

				if messageReact == nil {
					continue
				}
				if messageReact.Reactions == nil {
					continue
				}
				// Checks if the vote was successful and executes if so
				if messageReact.Reactions[0].Count < vote.GetVotesReq() {
					continue
				}

				// Tell the startvote command to wait for this command to finish before moving on
				entities.Mutex.Lock()
				inChanCreation = true
				entities.Mutex.Unlock()

				var (
					message discordgo.Message
					author  discordgo.User
				)

				// Fixes role name bugs
				role := strings.Replace(strings.TrimSpace(vote.GetChannel()), " ", "-", -1)
				role = strings.Replace(role, "--", "-", -1)
				reg, err := regexp.Compile("[^a-zA-Z0-9-]+")
				if err != nil {
					common.LogError(s, guildSettings.BotLog, err)
					continue
				}
				role = reg.ReplaceAllString(role, "")

				// Removes all hyphen prefixes and suffixes because discord cannot handle them
				for strings.HasPrefix(role, "-") || strings.HasSuffix(role, "-") {
					role = strings.TrimPrefix(role, "-")
					role = strings.TrimSuffix(role, "-")
				}

				vote.SetChannel(role)

				// Deletes vote
				err = db.SetGuildVoteInfoChannel(guild.ID, messID, vote, true)
				if err != nil {
					log.Println(err)
				}

				// Create command
				author.ID = s.State.User.ID
				message.ChannelID = vote.GetMessageReact().ChannelID
				message.GuildID = guild.ID
				message.Author = &author
				if vote.GetCategory() == "" {
					message.Content = fmt.Sprintf("%screate %s %s %s", guildSettings.GetPrefix(), vote.GetChannel(), vote.GetChannelType(), vote.GetDescription())
				} else {
					message.Content = fmt.Sprintf("%screate %s %s %s %s", guildSettings.GetPrefix(), vote.GetChannel(), vote.GetChannelType(), vote.GetCategory(), vote.GetDescription())
				}

				createChannelCommand(s, &message)

				time.Sleep(500 * time.Millisecond)

				// Sortroles command if optin, airing or temp
				if vote.GetChannelType() != "general" {
					message.Content = guildSettings.GetPrefix() + "sortroles"
					sortRolesCommand(s, &message)
				}

				time.Sleep(500 * time.Millisecond)

				// Sortcategory command if category exists and it's not temp
				if vote.GetCategory() != "" && vote.GetChannelType() != "temp" && vote.GetChannelType() != "temporary" {
					message.Content = guildSettings.GetPrefix() + "sortcategory " + vote.GetCategory()
					sortCategoryCommand(s, &message)
				}

				if vote.GetChannelType() == "general" {
					_, err = s.ChannelMessageSend(messageReact.ChannelID, fmt.Sprintf("Channel `%s` was successfuly created!", vote.GetChannel()))
					if err != nil {
						common.LogError(s, guildSettings.BotLog, err)
						entities.Mutex.Lock()
						inChanCreation = false
						entities.Mutex.Unlock()
						continue
					}
				} else {
					_, err = s.ChannelMessageSend(messageReact.ChannelID, fmt.Sprintf("Channel `%s` was successfully created! Those that have voted were given the role. Use `%sjoin %s` if you do not have it.", vote.GetChannel(), guildSettings.GetPrefix(), role))
					if err != nil {
						common.LogError(s, guildSettings.BotLog, err)
						entities.Mutex.Lock()
						inChanCreation = false
						entities.Mutex.Unlock()
						continue
					}
				}

				time.Sleep(200 * time.Millisecond)

				if vote.GetChannelType() != "general" {
					roles, err := s.GuildRoles(guild.ID)
					if err != nil {
						common.LogError(s, guildSettings.BotLog, err)
						entities.Mutex.Lock()
						inChanCreation = false
						entities.Mutex.Unlock()
						continue
					}
					for i := 0; i < len(roles); i++ {
						if roles[i].Name == role {
							role = roles[i].ID
							break
						}
					}

					// Gets the users who voted and gives them the role
					users, err := s.MessageReactions(vote.GetMessageReact().ChannelID, vote.GetMessageReact().ID, "👍", 100)
					if err != nil {
						common.LogError(s, guildSettings.BotLog, err)
						entities.Mutex.Lock()
						inChanCreation = false
						entities.Mutex.Unlock()
						continue
					}

					for i := 0; i < len(users); i++ {
						if users[i].ID == s.State.User.ID {
							continue
						}

						err := s.GuildMemberRoleAdd(guild.ID, users[i].ID, role)
						if err != nil {
							common.LogError(s, guildSettings.BotLog, err)
							continue
						}
					}
				}

				entities.Mutex.Lock()
				inChanCreation = false
				entities.Mutex.Unlock()

				if vote.GetChannelType() == "temp" || vote.GetChannelType() == "temporary" {
					if guildSettings.BotLog == (entities.Cha{}) {
						continue
					}
					if guildSettings.BotLog.GetID() != "" {
						continue
					}
					_, err = s.ChannelMessageSend(guildSettings.BotLog.GetID(), fmt.Sprintf("Temp channel `%s` has been created from a vote by user %s#%s.", vote.GetChannel(), vote.GetUser().Username, vote.GetUser().Discriminator))
					if err != nil {
						common.LogError(s, guildSettings.BotLog, err)
						continue
					}
					continue
				}
			}

			cha, err := s.GuildChannels(guild.ID)
			if err != nil {
				continue
			}

			guildTempCha := db.GetGuildTempChannels(guild.ID)

			for k, v := range guildTempCha {
				if v == nil {
					continue
				}

				for i := 0; i < len(cha); i++ {
					if cha[i].Name == v.GetRoleName() {
						mess, err := s.ChannelMessages(cha[i].ID, 1, "", "", "")
						if err != nil {
							common.LogError(s, guildSettings.BotLog, err)
							continue
						}

						// Fetches the properly parsed timestamp from discord, else uses channel creation
						if len(mess) != 0 {
							timestamp, err = mess[0].Timestamp.Parse()
							if err != nil {
								common.LogError(s, guildSettings.BotLog, err)
								continue
							}
						} else {
							timestamp = v.GetCreationDate()
						}

						// Adds how long before last message for channel to be deletes
						timestamp = timestamp.Add(time.Hour * 3)
						t := time.Now()

						// Calculates if it's time to remove
						difference := t.Sub(timestamp)
						if difference > 0 {

							if guildSettings.BotLog != (entities.Cha{}) {
								if guildSettings.BotLog.GetID() != "" {
									_, err = s.ChannelMessageSend(guildSettings.BotLog.GetID(), fmt.Sprintf("Temp channel `%v` has been deleted due to being inactive for 3 hours.", cha[i].Name))
									if err != nil {
										common.LogError(s, guildSettings.BotLog, err)
										continue
									}
								}
							}

							// Deletes channel and role
							_, err := s.ChannelDelete(cha[i].ID)
							if err != nil {
								common.LogError(s, guildSettings.BotLog, err)
								continue
							}

							err = s.GuildRoleDelete(guild.ID, k)
							if err != nil {
								common.LogError(s, guildSettings.BotLog, err)
								continue
							}

							err = db.SetGuildTempChannel(guild.ID, k, v, true)
							if err != nil {
								log.Println(err)
							}
						}
					}
				}
			}
		}
	}
}

func init() {
	Add(&Command{
		Execute: startVoteCommand,
		Trigger: "startvote",
		Desc:    "Starts a vote for channel creation [VOTE]",
		Module:  "channel",
	})
}
