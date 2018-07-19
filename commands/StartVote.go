package commands

import (
	"fmt"
	"strings"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
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

//Handles mod started channel creation vote
func ChannelVoteHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

		//Pulls info on message author
		mem, err := s.State.Member(config.ServerID, m.Author.ID)
		if err != nil {
			mem, err = s.GuildMember(config.ServerID, m.Author.ID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}

		//Puts the command to lowercase
		messageLowercase := strings.ToLower(m.Content)

		//Checks if user has permissions and whether the BotPrefix was used
		if strings.HasPrefix(messageLowercase, config.BotPrefix) {
			if misc.HasPermissions(mem) {
				if strings.HasPrefix(messageLowercase, config.BotPrefix+"startvote ") && (messageLowercase != (config.BotPrefix + "startvote")) {

					if m.Author.ID == config.BotID {
						return
					}

					// Initializes necessary variables and sets some starting values for them
					var (
						peopleNum     = 7
						channelType   = "opt-in"
						channelName   string
						description   string
						categoryIndex int
						messageID     = m.ID
					)

					// Splits the original message by spaces so we can parse for parameters
					messageSplit := strings.Split(messageLowercase, " ")

					// Checks if the second parameter is a number (required number of people parameter) and assigns that number
					// It also parses for second paremeter if that is true
					// Otherwise it checks for channel type parameter directly
					if num, err := strconv.Atoi(messageSplit[1]); err == nil {

						// Assigns the num to peopleNum
						peopleNum = num

						// If it's optin do nothing since that's the default
						if messageSplit[2] == "optin" || messageSplit[2] == "opt-in" {

							// If it's airing assign that string value to channelType
						} else if messageSplit[2] == "airing" {

							channelType = "airing"

							// If it's general then assign that channel type
						} else if messageSplit[2] == "general" {

							channelType = "general"

							// Else prints an error message
						}

						// Checks if the string is 18 in length and a number. If true it converts to a number and assigns as categoryID
						for i := 0; i < len(messageSplit); i++ {
							if len(messageSplit[i]) == 18 {
								if _, err := strconv.Atoi(messageSplit[i]); err == nil {

									categoryIndex = i
								}
							}
						}

					} else {

						// If it's optin do nothing since that's the default
						if messageSplit[1] == "optin" || messageSplit[1] == "opt-in" {

							// If it's airing assign that string value to channelType
						} else if messageSplit[1] == "airing" {

							channelType = "airing"

							// If it's general then assign that channel type
						} else if messageSplit[1] == "general" {

							channelType = "general"

							// Else prints an error message
						}

						// Checks if the string is 18 in length and a number. If true it converts to a number and assigns as categoryID
						for i := 0; i < len(messageSplit); i++ {
							if len(messageSplit[i]) == 18 {
								if _, err := strconv.Atoi(messageSplit[i]); err == nil {

									categoryIndex = i
								}
							}
						}
					}

					// If a description and categoryID exist , it parses the description
					if len(messageSplit) > categoryIndex && categoryIndex != 0 {
						for i := categoryIndex + 1; i < len(messageSplit); i++ {

							// Simple check so it doesn't add an unnecessary space at start
							if description != "" {

								description = description + " " + messageSplit[i]
							} else {

								description = messageSplit[i]
							}
						}
					}

					// After parsing all of the possible parameters we finally fetch the channel name
					channelName = messageLowercase
					peopleNumStr := strconv.Itoa(peopleNum)
					channelName = strings.Replace(channelName, messageSplit[0], "", -1)
					if messageSplit[1] == peopleNumStr {
						channelName = strings.Replace(channelName, peopleNumStr, "", 1)
					}
					channelName = strings.Replace(channelName, channelType, "", -1)
					channelName = strings.Replace(channelName, messageSplit[categoryIndex], "", -1)
					channelName = strings.Replace(channelName, description, "", -1)

					// If the name doesn't exist, print an error
					if channelName == "" {

						// Prints error
						failure := "Error: Channel name not parsed properly. Please use `" + config.BotPrefix + "startvote " +
							"OPTIONAL[votes required] OPTIONAL[type] [name] OPTIONAL[categoryID] OPTIONAL[description]"
						_, err = s.ChannelMessageSend(m.ChannelID, failure)
						if err != nil {

							fmt.Println("Error:", err)
						}
					}

					peopleNum = peopleNum + 1
					peopleNumStr = strconv.Itoa(peopleNum)
					messageReact, err := s.ChannelMessageSend(m.ChannelID, peopleNumStr+" thumbs up reacts on this message will create `"+channelName+"`. Time limit is 30 hours")
					if err != nil {

						fmt.Println("Error", err)
					}

					if messageReact != nil {

						err = s.MessageReactionAdd(messageReact.ChannelID, messageReact.ID, "👍")
						if err != nil {

							fmt.Println("Error:", err)
						}

						messageReact, err = s.ChannelMessage(messageReact.ChannelID, messageReact.ID)
						if err != nil {

							fmt.Println("Error:", err)
						}
					}

					//Saves the current time from which it'll start counting the 30 hours later
					t := time.Now()

					// Reads ongoing votes from VoteInfo.json
					VoteInfoRead()

					var temp VoteInfo

					MapMutex.Lock()

					// Saves the date of removal in separate variable and then adds 30 hours to it
					thirtyHours := time.Hour * 30
					dateRemoval := t.Add(thirtyHours)

					// Assigns values to VoteInfoMap so it can be written to storage
					temp.Date = dateRemoval
					temp.Channel = channelName
					temp.Description = description
					if len(messageSplit[categoryIndex]) == 18 {
						if _, err := strconv.Atoi(messageSplit[categoryIndex]); err == nil {

							temp.Category = messageSplit[categoryIndex]
						}
					}
					temp.ChannelType = channelType
					temp.VotesReq = peopleNum
					temp.MessageReact = messageReact

					VoteInfoMap[messageID] = &temp

					MapMutex.Unlock()

					// Writes to storage
					VoteInfoWrite(VoteInfoMap)
				}
			}
		}
	}
}

// Checks if the message has enough reacts every 10 seconds, and stops if it's over the time limit
func ChannelVoteTimer(s *discordgo.Session, e *discordgo.Ready) {

	for range time.NewTicker(10 * time.Second).C {

		// Reads ongoing votes from VoteInfo.json
		VoteInfoRead()

		for k := range VoteInfoMap {

			// Reads from storage
			VoteInfoRead()

			// Saves current time
			t := time.Now()

			// Checks message constantly
			messageReact, err := s.ChannelMessage(VoteInfoMap[k].MessageReact.ChannelID, VoteInfoMap[k].MessageReact.ID)
			if err != nil {

				fmt.Println("Error:", err)
			}

			// Calculates if it's time to remove
			difference := t.Sub(VoteInfoMap[k].Date)
			if difference > 0 {

				if messageReact != nil {

					numStr := strconv.Itoa(VoteInfoMap[k].VotesReq - 1)
					_, err = s.ChannelMessageSend(messageReact.ChannelID, "Channel vote has ended. `"+VoteInfoMap[k].Channel+"` has failed to "+
						"gather the necessary "+numStr+" votes.")
					if err != nil {

						fmt.Println("Error:", err)
					}
				}

				//Deletes from map and storage

				MapMutex.Lock()

				// Deletes the vote from memory
				delete(VoteInfoMap, k)

				MapMutex.Unlock()

				// Writes to storage
				VoteInfoWrite(VoteInfoMap)

				continue
			}

			// A few checks so it doesn't crash the program
			if messageReact != nil {
				if messageReact.Reactions != nil {
					if messageReact.Reactions[0].Count >= VoteInfoMap[k].VotesReq {

						// Executes create command
						success := config.BotPrefix + "create" + VoteInfoMap[k].Channel + " " + VoteInfoMap[k].ChannelType +
							" " + VoteInfoMap[k].Category + " " + VoteInfoMap[k].Description
						message, err := s.ChannelMessageSend(VoteInfoMap[k].MessageReact.ChannelID, success)
						if err != nil {

							fmt.Println("Error:", err)
						}
						s.ChannelMessageDelete(message.ChannelID, message.ID)

						time.Sleep(2 * time.Second)

						// Executes sortroles if optin or airing

						if VoteInfoMap[k].ChannelType == "opt-in" || VoteInfoMap[k].ChannelType == "airing" {

							success = config.BotPrefix + "sortroles"
							message, err = s.ChannelMessageSend(config.BotLogID, success)
							if err != nil {

								fmt.Println("Error:", err)
							}
							s.ChannelMessageDelete(message.ChannelID, message.ID)
						}

						time.Sleep(2 * time.Second)

						// Executes sortcategory if category exists
						if VoteInfoMap[k].Category != "" && len(VoteInfoMap[k].Category) == 18 {

							success = config.BotPrefix + "sortcategory " + VoteInfoMap[k].Category
							message, err = s.ChannelMessageSend(config.BotLogID, success)
							if err != nil {

								fmt.Println("Errorr:", err)
							}
							s.ChannelMessageDelete(message.ChannelID, message.ID)
						}

						//Deletes from map and storage
						// Reads from storage
						VoteInfoRead()

						MapMutex.Lock()

						// Deletes the vote from memory
						delete(VoteInfoMap, k)

						MapMutex.Unlock()

						// Writes to storage
						VoteInfoWrite(VoteInfoMap)
					}
				}
			}
		}
	}
}

//Reads vote info from voteInfo.json
func VoteInfoRead() {

	MapMutex.Lock()

	// Reads all vote channel info voteInfo.json file and puts them in VoteInfoMap as bytes
	voteInfoByte, err := ioutil.ReadFile("database/voteInfo.json")
	if err != nil {

		//fmt.Println(err)
	}

	// Takes all of the vote channels from voteInfo.json from byte and puts them into the VoteInfo map
	err = json.Unmarshal(voteInfoByte, &VoteInfoMap)
	if err != nil {

		//fmt.Println(err)
	}

	MapMutex.Unlock()
}

// Writes vote info to voteInfo.json
func VoteInfoWrite(info map[string]*VoteInfo) {

	MapMutex.Lock()

	//Turns info slice into byte ready to be pushed to file
	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	//Writes to file
	err = ioutil.WriteFile("database/voteInfo.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}

	MapMutex.Unlock()
}
