package commands

import (
	"strings"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

var (
	VoteInfoMap = make(map[string]*VoteInfo)
	TempChaMap  = make(map[string]*TempChaInfo)
)

// VoteInfo is the in memory storage of each vote channel's info
type VoteInfo struct {
	Date         time.Time          `json:"Date"`
	Channel      string             `json:"Channel"`
	ChannelType  string             `json:"ChannelType"`
	Category     string             `json:"Category,omitempty"`
	Description  string             `json:"Description,omitempty"`
	VotesReq     int                `json:"VotesReq"`
	MessageReact *discordgo.Message `json:"MessageReact"`
	User		 *discordgo.User	`json:"User"`
}

type TempChaInfo struct {
	CreationDate	time.Time		`json:"CreationDate"`
	RoleName		string			`json:"RoleName"`
	Elevated		bool			`json:"Elevated"`
}

// Starts a 30 hour vote in a channel with parameters that automatically creates a spoiler channel if the vote passes
func startVoteCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		peopleNum = 		7
		descriptionSlice 	[]string
		voteChannel 		channel
		controlNumVote =	0
		controlNumUser =	0
		controlNum =		0
	)

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if hasElevatedPermissions(s, m.Author) {
		if len(commandStrings) == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"startvote OPTIONAL[votes required] [name] OPTIONAL[type] OPTIONAL[categoryID] + OPTIONAL[description]`")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
				if err != nil {
					return
				}
				return
			}
			return
		}

		command := strings.Replace(messageLowercase, config.BotPrefix+"startvote ", "", 1)
		commandStrings = strings.Split(command, " ")

		// Checks if [category] and [type] exist and assigns them if they do and removes them from slice
		for i := 0; i < len(commandStrings); i++ {
			_, err := strconv.Atoi(commandStrings[i])
			if len(commandStrings[i]) >= 17 && err == nil {
				voteChannel.Category = commandStrings[i]
				command = strings.Replace(command, commandStrings[i], "", -1)
			}

			if commandStrings[i] == "airing" ||
				commandStrings[i] == "general" ||
				commandStrings[i] == "opt-in" ||
				commandStrings[i] == "optin" ||
				commandStrings[i] == "temp" ||
				commandStrings[i] == "temporary" {

				voteChannel.Type = commandStrings[i]
				command = strings.Replace(command, commandStrings[i], "", -1)
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
		if blank == true {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Channel name not parsed properly. Please use `" + config.BotPrefix + "startvote "+
				"OPTIONAL[votes required] [name] OPTIONAL[type] OPTIONAL[categoryID] + OPTIONAL[description]`")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
				if err != nil {
					return
				}
				return
			}
			return
		}
	} else {
		if len(commandStrings) == 1 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"startvote [name]`")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
				if err != nil {
					return
				}
				return
			}
			return
		}

		// Pulls up all current server channels and checks if it exists in UserTempCha.json. If not it deletes it from storage
		cha, err := s.GuildChannels(config.ServerID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err)
		}
		for k, v := range TempChaMap {
			exists := false
			for i := 0; i < len(cha); i++ {
				if cha[i].Name == v.RoleName {
					exists = true
					break
				}
			}
			if !exists {
				MapMutex.Lock()
				delete(TempChaMap, k)
				MapMutex.Unlock()

				TempChaWrite(TempChaMap)
			}

			if v.Elevated == false {
				controlNumUser++
			}
		}

		// Prints error if the user temp channel cap (3) has been reached
		if controlNumUser > 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Maximum number of user made temp channels(3) has been reached." +
				" Please contact a mod for a new temp channel or wait for the other three to run out.")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
				if err != nil {
					return
				}
				return
			}
			return
		}

		// Checks if there are any current temp channel votes made by users and already created channels that have reached the cap and prints error if there are
		for _, v := range VoteInfoMap {
			if !hasElevatedPermissions(s, v.User) {
				controlNumVote++
			}
		}
		controlNum = controlNumVote + controlNumUser
		if controlNum > 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are already ongoing user temp votes that breach the cap(3) together with already created temp channels. " +
				"Please contact a mod for a new temp channel or wait for the votes or temp channels to run out.")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
				if err != nil {
					return
				}
				return
			}
			return
		}

		// Initializes default variables
		name := strings.Replace(messageLowercase, config.BotPrefix + "startvote ", "", -1)
		voteChannel.Category = "486823979764678657"
		voteChannel.Type = "temp"
		voteChannel.Description = "Temporary channel for " + name + ". Will be deleted 3 hours after no message has been sent."
		peopleNum = 3

		// Fixes role name bugs
		role := strings.Replace(strings.TrimSpace(name), " ", "-", -1)
		role = strings.Replace(role, "--", "-", -1)
		voteChannel.Name = role
	}

	peopleNum = peopleNum + 1
	peopleNumStr := strconv.Itoa(peopleNum)
	messageReact, err := s.ChannelMessageSend(m.ChannelID, peopleNumStr+" thumbs up reacts on this message will create `"+voteChannel.Name+"`. Time limit is 30 hours")
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	if messageReact == nil {
		return
	}

	err = s.MessageReactionAdd(messageReact.ChannelID, messageReact.ID, "👍")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
		if err != nil {
			return
		}
	}
	messageReact, err = s.ChannelMessage(messageReact.ChannelID, messageReact.ID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	t := time.Now()

	var temp VoteInfo

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

	MapMutex.Lock()
	VoteInfoMap[m.ID] = &temp
	MapMutex.Unlock()

	// Writes to storage
	VoteInfoWrite(VoteInfoMap)
}

// Checks if the message has enough reacts every 10 seconds, and stops if it's over the time limit
func ChannelVoteTimer(s *discordgo.Session, e *discordgo.Ready) {

	var (
		//roleName string
		timestamp time.Time
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {

				fmt.Println(err)
				fmt.Println(rec)
			}
		}
	}()

	for range time.NewTicker(10 * time.Second).C {
		for k := range VoteInfoMap {

			t := time.Now()

			// Updates message
			messageReact, err := s.ChannelMessage(VoteInfoMap[k].MessageReact.ChannelID, VoteInfoMap[k].MessageReact.ID)
			if err != nil {
				return
			}

			// Calculates if it's time to remove
			difference := t.Sub(VoteInfoMap[k].Date)
			if difference > 0 {
				if messageReact == nil {
					continue
				}

				// Fixes role name bugs
				role := strings.Replace(strings.TrimSpace(VoteInfoMap[k].Channel), " ", "-", -1)
				role = strings.Replace(role, "--", "-", -1)

				numStr := strconv.Itoa(VoteInfoMap[k].VotesReq - 1)
				_, err = s.ChannelMessageSend(messageReact.ChannelID, "Channel vote has ended. `" + VoteInfoMap[k].Channel + "` has failed to "+
					"gather the necessary "+ numStr+ " votes.")
				if err != nil {
					_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
					if err != nil {
						return
					}
				}

				// Deletes the vote from memory
				MapMutex.Lock()
				delete(VoteInfoMap, k)
				MapMutex.Unlock()

				// Writes to storage
				VoteInfoWrite(VoteInfoMap)
				continue
			}

			if messageReact == nil {

				continue
			}
			if messageReact.Reactions == nil {

				continue
			}
			// Checks if the vote was successful and executes if so
			if messageReact.Reactions[0].Count < VoteInfoMap[k].VotesReq {
				continue
			}

			var (
				message discordgo.Message
				author  discordgo.User
			)

			// Fixes role name bugs
			role := strings.Replace(strings.TrimSpace(VoteInfoMap[k].Channel), " ", "-", -1)
			role = strings.Replace(role, "--", "-", -1)

			// Removes all hyphen prefixes and suffixes because discord cannot handle them
			for strings.HasPrefix(role, "-") || strings.HasSuffix(role, "-") {
				role = strings.TrimPrefix(role, "-")
				role = strings.TrimSuffix(role, "-")
			}

			VoteInfoMap[k].Channel = role

			// Create command
			author.ID = s.State.User.ID
			message.ID = VoteInfoMap[k].MessageReact.ChannelID
			message.Author = &author
			message.Content = config.BotPrefix + "create " + VoteInfoMap[k].Channel + " " + VoteInfoMap[k].ChannelType +
				" " + VoteInfoMap[k].Category + " " + VoteInfoMap[k].Description
			createChannelCommand(s, &message)

			time.Sleep(200 * time.Millisecond)

			// Sortroles command if optin, airing or temp
			if VoteInfoMap[k].ChannelType != "general" {
				message.Content = config.BotPrefix + "sortroles"
				sortRolesCommand(s, &message)
			}

			time.Sleep(500 * time.Millisecond)

			// Sortcategory command if category exists and it's not temp
			if VoteInfoMap[k].Category != "" && VoteInfoMap[k].ChannelType != "temp" || VoteInfoMap[k].ChannelType != "temporary" {
				message.Content = config.BotPrefix + "sortcategory " + VoteInfoMap[k].Category
				sortCategoryCommand(s, &message)
			}

			if VoteInfoMap[k].ChannelType != "temp" {
				_, err = s.ChannelMessageSend(messageReact.ChannelID, "Channel `" + VoteInfoMap[k].Channel + "` was successfully created! Those that have voted were given the role. Use `"+
					config.BotPrefix+ "join "+ role + "` until reaction join has been set if you do not have it.")
				if err != nil {
					_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
					if err != nil {
						return
					}
					return
				}
			} else {
				_, err = s.ChannelMessageSend(messageReact.ChannelID, "Channel `" + VoteInfoMap[k].Channel + "` was successfully created! Those that have voted were given the role. Use `"+
					config.BotPrefix+ "join "+ role + "` otherwise.")
				if err != nil {
					_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
					if err != nil {
						return
					}
					return
				}
			}

			time.Sleep(200 * time.Millisecond)

			roles, err := s.GuildRoles(config.ServerID)
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
				if err != nil {
					return
				}
				return
			}
			for i := 0; i < len(roles); i++ {
				if roles[i].Name == role {
					role = roles[i].ID
					//roleName = roles[i].Name
					break
				}
			}

			// Gets the users who voted and gives them the role
			users, err := s.MessageReactions(VoteInfoMap[k].MessageReact.ChannelID, VoteInfoMap[k].MessageReact.ID, "👍", 100)
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
				if err != nil {
					return
				}
				return
			}

			for i := 0; i < len(users); i++ {
				if users[i].ID == config.BotID {
					continue
				}

				err := s.GuildMemberRoleAdd(config.ServerID, users[i].ID, role)
				if err != nil {
					_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
					if err != nil {
						return
					}
					return
				}
			}

			// Deletes the vote from memory
			MapMutex.Lock()
			delete(VoteInfoMap, k)
			MapMutex.Unlock()

			// Writes to storage
			VoteInfoWrite(VoteInfoMap)
		}

		cha, err := s.GuildChannels(config.ServerID)
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {
				return
			}
			return
		}

		for k, v := range TempChaMap {
			for i := 0; i < len(cha); i++ {
				if cha[i].Name == v.RoleName {
					mess, err := s.ChannelMessages(cha[i].ID, 1, "", "", "")
					if err != nil {
						_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
						if err != nil {
							return
						}
						return
					}

					// Fetches the properly parsed timestamp from discord, else uses channel creation
					if len(mess) != 0 {
						timestamp, err = mess[0].Timestamp.Parse()
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
							if err != nil {
								return
							}
							return
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
						// Deletes channel and role
						_, err := s.ChannelDelete(cha[i].ID)
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
							if err != nil {
								return
							}
							return
						}

						err = s.GuildRoleDelete(config.ServerID, k)
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
							if err != nil {
								return
							}
							return
						}

						MapMutex.Lock()
						delete(TempChaMap, k)
						MapMutex.Unlock()

						TempChaWrite(TempChaMap)
					}
				}
			}
		}
	}
}

// Reads vote info from voteInfo.json
func VoteInfoRead() {

	// Reads all vote channel info voteInfo.json file and puts them in VoteInfoMap as bytes
	voteInfoByte, err := ioutil.ReadFile("database/voteInfo.json")
	if err != nil {
		return
	}

	// Takes all of the vote channels from voteInfo.json from byte and puts them into the VoteInfo map
	MapMutex.Lock()
	err = json.Unmarshal(voteInfoByte, &VoteInfoMap)
	if err != nil {
		MapMutex.Unlock()
		return
	}
	MapMutex.Unlock()
}

// Writes vote info to voteInfo.json
func VoteInfoWrite(info map[string]*VoteInfo) {

	// Turns info slice into byte ready to be pushed to file
	MapMutex.Lock()
	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		MapMutex.Unlock()
		return
	}
	MapMutex.Unlock()

	//Writes to file
	err = ioutil.WriteFile("database/voteInfo.json", MarshaledStruct, 0644)
	if err != nil {
		return
	}
}

// Reads temp channels from TempCha.json
func TempChaRead() {

	// Reads all user temp channel info userTempCha.json file and puts them in tempChaByte as bytes
	tempChaByte, err := ioutil.ReadFile("database/tempCha.json")
	if err != nil {
		return
	}

	// Takes all of the user temp channels from tempCha.json from byte and puts them into the tempChaMap map
	MapMutex.Lock()
	err = json.Unmarshal(tempChaByte, &TempChaMap)
	if err != nil {
		MapMutex.Unlock()
		return
	}
	MapMutex.Unlock()
}

// Writes temp cha info to tempCha.json
func TempChaWrite(info map[string]*TempChaInfo) {

	// Turns info map into byte ready to be pushed to file
	MapMutex.Lock()
	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		MapMutex.Unlock()
		return
	}
	MapMutex.Unlock()

	// Writes to file
	err = ioutil.WriteFile("database/tempCha.json", MarshaledStruct, 0644)
	if err != nil {
		return
	}
}

func init() {
	add(&command{
		execute:  startVoteCommand,
		trigger:  "startvote",
		desc:     "Starts a vote for channel creation with parameters.",
		elevated: false,
	})
}