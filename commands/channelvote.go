package commands

import (
	"strings"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"ZeroTsu/misc"
)

var VoteInfoMap = make(map[string]*VoteInfo)

// VoteInfo is the in memory storage of each vote channel's info
type VoteInfo struct {
	Date         time.Time          `json:"Date"`
	Channel      string             `json:"Channel"`
	ChannelType  string             `json:"ChannelType"`
	Category     string             `json:"Category,omitempty"`
	Description  string             `json:"Description,omitempty"`
	VotesReq     int                `json:"VotesReq"`
	MessageReact *discordgo.Message `json:"MessageReact"`
}

// Starts a 30 hour vote in a channel with parameters that automatically creates a spoiler channel if the vote passes
func startVoteCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		peopleNum= 7
		descriptionSlice []string

		voteChannel channel
	)

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

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
			commandStrings[i] == "optin" {

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
		command = strings.Replace(command, commandStrings[0], "", 1)
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

	MapMutex.Lock()

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

	VoteInfoMap[m.ID] = &temp

	MapMutex.Unlock()

	// Writes to storage
	VoteInfoWrite(VoteInfoMap)
}

// Checks if the message has enough reacts every 10 seconds, and stops if it's over the time limit
func ChannelVoteTimer(s *discordgo.Session, e *discordgo.Ready) {

	for range time.NewTicker(10 * time.Second).C {
		for k := range VoteInfoMap {

			// Saves program from panic and continues running normally without executing the command if it happens
			defer func() {
				if rec := recover(); rec != nil {
					_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
					if err != nil {

						return
					}
				}
			}()

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

				return
			}

			var (
				message discordgo.Message
				author  discordgo.User
			)

			// Create command
			author.ID = s.State.User.ID
			message.ID = VoteInfoMap[k].MessageReact.ChannelID
			message.Author = &author
			message.Content = config.BotPrefix + "create" + VoteInfoMap[k].Channel + " " + VoteInfoMap[k].ChannelType +
				" " + VoteInfoMap[k].Category + " " + VoteInfoMap[k].Description
			createChannelCommand(s, &message)

			time.Sleep(500 * time.Millisecond)

			// Sortroles command if optin or airing
			if VoteInfoMap[k].ChannelType != "general" {

				message.Content = config.BotPrefix + "sortroles"
				sortRolesCommand(s, &message)
			}

			time.Sleep(2 * time.Second)

			// Sortcategory command if category exists
			if VoteInfoMap[k].Category != "" {

				message.Content = config.BotPrefix + "sortcategory " + VoteInfoMap[k].Category
				sortCategoryCommand(s, &message)

			}

			// Deletes the vote from memory
			MapMutex.Lock()
			delete(VoteInfoMap, k)
			MapMutex.Unlock()

			// Writes to storage
			VoteInfoWrite(VoteInfoMap)

			_, err = s.ChannelMessageSend(config.BotLogID, "Channel `"+VoteInfoMap[k].Channel+"` was successfully created!")
			if err != nil {

				_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
				if err != nil {

					return
				}
				return
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

		return
	}
	MapMutex.Unlock()

	//Writes to file
	err = ioutil.WriteFile("database/voteInfo.json", MarshaledStruct, 0644)
	if err != nil {

		return
	}
}

func init() {
	add(&command{
		execute:  startVoteCommand,
		trigger:  "startvote",
		desc:     "Starts a vote for channel creation with parameters.",
		elevated: true,
	})
}