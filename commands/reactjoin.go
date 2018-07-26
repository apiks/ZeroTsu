package commands

import (
	"fmt"
	"regexp"
	"strings"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
	"github.com/r-anime/ZeroTsu/config"
)

var (
	reactChannelJoinMap = make(map[string]*ReactChannelJoinStruct)
	EmojiRoleMap        = make(map[string][]string)
	MapMutex            sync.Mutex
)

type ReactChannelJoinStruct struct {
	MessageID string                `json:"messageID"`
	RoleEmoji []map[string][]string `json:"roleEmoji"`
}

// Gives a specific role to a user if they react
func ReactJoinHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	// Checks if a react channel join is set for that specific message and emoji and continues if true
	if reactChannelJoinMap[r.MessageID] == nil {

		return
	}
	// Checks if it's the correct message and emoji before going down
	if reactChannelJoinMap[r.MessageID].MessageID != r.MessageID {

		return
	}

	// Puts the react API name to lowercase so it is valid with the storage emoji name
	reactLowercase := strings.ToLower(r.Emoji.APIName())

	for p := 0; p < len(reactChannelJoinMap[r.MessageID].RoleEmoji); p++ {
		for role := range reactChannelJoinMap[r.MessageID].RoleEmoji[p] {
			for _, emote := range reactChannelJoinMap[r.MessageID].RoleEmoji[p][role] {
				if reactLowercase != emote {

					continue
				}

				if len(role) >= 17 {
					if _, err := strconv.ParseInt(role, 10, 64); err == nil {

						// Gives the role
						err := s.GuildMemberRoleAdd(config.ServerID, r.UserID, role)
						if err != nil {

							fmt.Println("Error:", err)
						}
					}
				} else {

					// Pulls all of the server roles
					roles, err := s.GuildRoles(config.ServerID)
					if err != nil {

						fmt.Println("Error:", err)
					}

					// Iterates through all of the server roles and gives the role to the user if the set role exists
					for i := 0; i < len(roles); i++ {
						if roles[i].Name == role {

							// Gives the role
							err := s.GuildMemberRoleAdd(config.ServerID, r.UserID, roles[i].ID)
							if err != nil {

								fmt.Println("Error:", err)
							}
						}
					}
				}
			}
		}
	}
}

// Removes a role from user if they unreact
func ReactRemoveHandler(s *discordgo.Session, r *discordgo.MessageReactionRemove) {

	// Checks if a react channel join is set for that specific message and emoji and continues if true
	if reactChannelJoinMap[r.MessageID] == nil {

		return
	}
	// Checks if it's the correct message and emoji before going down
	if reactChannelJoinMap[r.MessageID].MessageID != r.MessageID {

		return
	}

	// Puts the react API name to lowercase so it is valid with the storage emoji name
	reactLowercase := strings.ToLower(r.Emoji.APIName())

	for p := 0; p < len(reactChannelJoinMap[r.MessageID].RoleEmoji); p++ {
		for role := range reactChannelJoinMap[r.MessageID].RoleEmoji[p] {
			for _, emote := range reactChannelJoinMap[r.MessageID].RoleEmoji[p][role] {
				if reactLowercase != emote {

					continue
				}

				if len(role) >= 17 {
					if _, err := strconv.ParseInt(role, 10, 64); err == nil {

						// Removes the role
						err := s.GuildMemberRoleRemove(config.ServerID, r.UserID, role)
						if err != nil {

							fmt.Println("Error:", err)
						}
					}
				} else {

					// Pulls all of the server roles
					roles, err := s.GuildRoles(config.ServerID)
					if err != nil {

						fmt.Println("Error:", err)
					}

					// Iterates through all of the server roles and removes the role from the user if the set role exists
					for i := 0; i < len(roles); i++ {
						if roles[i].Name == role {

							// Removes the role
							err := s.GuildMemberRoleRemove(config.ServerID, r.UserID, roles[i].ID)
							if err != nil {

								fmt.Println("Error:", err)
							}
						}
					}
				}
			}
		}
	}
}

// Sets react joins per specific message and emote
func setReactJoinCommand (s *discordgo.Session, m *discordgo.Message) {

	// Puts the command to lowercase
	messageLowercase := strings.ToLower(m.Content)

	// Splits the message parameters
	commandStrings := strings.SplitN(messageLowercase, " ", 4)

	if len(commandStrings) != 4 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Wrong amount of parameters. Please use `" + config.BotPrefix + "setreact [messageID] [emoji] [role]`")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}
	num, err := strconv.Atoi(commandStrings[1])
	if err != nil || num < 17 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid messageID.")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}

	// Checks if the role exists on the server
	var roleExists bool
	roles, err := s.GuildRoles(config.ServerID)
	if err != nil {

		fmt.Println("Error: ", err)
	}
	// Iterates through all of the roles and sets bool value to true if the role exists
	for j := 0; j < len(roles); j++ {

		role := strings.ToLower(roles[j].Name)
		if role == commandStrings[3] {
			roleExists = true
			break
		}
	}
	if roleExists == false {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid role.")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}

	// Parses if it's custom emoji or unicode emoji
	re := regexp.MustCompile("(?i)<:+([a-zA-Z]|[0-9])+:+[0-9]+>")
	emojiRegex := re.FindAllString(messageLowercase, 1)

	if emojiRegex != nil {

		// Fetches emoji API name
		re := regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
		emojiName := re.FindAllString(emojiRegex[0], 1)

		// Sets the data in memory to be ready for writing
		SaveReactJoin(commandStrings[1], commandStrings[3], emojiName[0])

		// Writes the data to storage
		ReactChannelJoinWrite(reactChannelJoinMap)

		// Reacts with the set emote if able
		err := s.MessageReactionAdd(m.ChannelID, commandStrings[1], emojiName[0])
		if err != nil {

			// Lots of checks because of PTSD
			message, err := s.ChannelMessage(m.ChannelID, commandStrings[1])
			if err != nil {
			} else if message.Reactions != nil {
				if len(message.Reactions) > 0 {
					if message.Reactions[0] != nil {
						if message.Reactions[0].Count == 20 {
							_, err = s.ChannelMessageSend(m.ChannelID, "Error: Reached the max reaction limit. (20)")
							if err != nil {

								fmt.Println("Error: ", err)
							}
						}
					}
				}
			}
		}
		_, err = s.ChannelMessageSend(m.ChannelID, "Success! React channel join set.")
		if err != nil {

			fmt.Println("Error:", err)
		}

		// Else it's a non-valid emoji or an unicode emoji (the latter preferably) and saves that
	} else {

		// Sets the data in memory to be ready for writing
		SaveReactJoin(commandStrings[1], commandStrings[3], commandStrings[2])

		// Writes the data to storage
		ReactChannelJoinWrite(reactChannelJoinMap)

		// Reacts with the set emote if able
		err := s.MessageReactionAdd(m.ChannelID, commandStrings[1], commandStrings[2])
		if err != nil {

			// Lots of checks because of PTSD
			message, err := s.ChannelMessage(m.ChannelID, commandStrings[1])
			if err != nil {
			} else if message != nil {
				if len(message.Reactions) > 0 {
					if message.Reactions[0] != nil {
						if message.Reactions[0].Count == 20 {
							_, err = s.ChannelMessageSend(m.ChannelID, "Error: Reached the max reaction limit. (20)")
							if err != nil {

								fmt.Println("Error:", err)
							}
						}
					}
				}
			}
		}
		_, err = s.ChannelMessageSend(m.ChannelID, "Success! React channel join set.")
		if err != nil {

			fmt.Println("Error:", err)
		}
	}
}

// Removes a previously set react join from memory and storage. Yes, it'll make your eyes bleed
func removeReactJoinCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		messageExists bool
		key           string
	)

	// Puts the command to lowercase
	messageLowercase := strings.ToLower(m.Content)

	// Splits the message parameters
	commandStrings := strings.SplitN(messageLowercase, " ", 3)

	if len(commandStrings) != 3 && len(commandStrings) != 2 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Wrong amount of parameters. Please use `"+config.BotPrefix+"removereact [messageID] [emoji]`")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}
	num, err := strconv.Atoi(commandStrings[1])
	if err != nil || num < 17 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid messageID.")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}
	if len(reactChannelJoinMap) == 0 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set react joins.")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}

	// Checks if the messageID already exists in the map
	for k := range reactChannelJoinMap {

		if commandStrings[1] == k {

			messageExists = true
			key = k
		}
	}
	if messageExists == false {

		_, err = s.ChannelMessageSend(m.ChannelID, "Error: No such messageID is set in storage")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}

	if len(commandStrings) == 2 {

		// Removes the entire message from the map and writes to storage
		misc.MapMutex.Lock()
		delete(reactChannelJoinMap, commandStrings[1])
		misc.MapMutex.Unlock()
		ReactChannelJoinWrite(reactChannelJoinMap)

		_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed entire message emoji react join.")
		if err != nil {

			fmt.Println("Error:", err)
		}

		return
	}

	if reactChannelJoinMap[key].RoleEmoji == nil {

		return
	}

	if len(commandStrings) == 3 {

		for p := 0; p < len(reactChannelJoinMap[key].RoleEmoji); p++ {
			for role := range reactChannelJoinMap[key].RoleEmoji[p] {
				for _, emote := range reactChannelJoinMap[key].RoleEmoji[p][role] {

					// Parses if it's custom emoji or unicode
					re := regexp.MustCompile("(?i)<:+([a-zA-Z]|[0-9])+:+[0-9]+>")
					emojiRegex := re.FindAllString(commandStrings[2], 1)

					// Second parser if it's custom emoji or unicode but for emoji API name instead
					reTwo := regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
					emojiRegexTwo := reTwo.FindAllString(commandStrings[2], 1)

					// Check whether emojiRegexTwo is empty or not to avoid index error
					var regexTwoBool = false
					if len(emojiRegexTwo) > 0 {
						if emote == emojiRegexTwo[0] {

							regexTwoBool = true
						}
					}

					// Checks if either regexTwoBool exists or the emote <-> commandStrings[2] are the same
					if emote == commandStrings[2] || regexTwoBool == true {
						if len(reactChannelJoinMap[key].RoleEmoji[p][role]) == 1 && len(reactChannelJoinMap[key].RoleEmoji[p]) == 1 {
							if emojiRegex != nil || emojiRegexTwo != nil {

								var emojiName []string

								// Fetches emoji API name
								re = regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
								if emojiRegex != nil {

									emojiName = re.FindAllString(emojiRegex[0], 1)

								} else if emojiRegexTwo != nil {

									emojiName = re.FindAllString(emojiRegexTwo[0], 1)
								}

								// Checks if the emoji in the map is the same as the one in the command
								if emojiName[0] == emote {

									// Removes the entire message from the map and writes to storage
									delete(reactChannelJoinMap, commandStrings[1])
									ReactChannelJoinWrite(reactChannelJoinMap)

									_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed Emoji react join from message.")
									if err != nil {

										fmt.Println("Error:", err)
									}

									break
								} else {

									_, err = s.ChannelMessageSend(m.ChannelID, "Error: No such emoji for that messageID")
									if err != nil {

										fmt.Println("Error:", err)
									}
								}

								break

								// If it's an unicode message it removes it with this
							} else {

								// Checks if the emoji in the map is the same as the one in the command
								if commandStrings[2] == emote {

									// Removes the entire message from the map
									delete(reactChannelJoinMap, commandStrings[1])
									ReactChannelJoinWrite(reactChannelJoinMap)

									_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed Emoji react join from message.")
									if err != nil {

										fmt.Println("Error:", err)
									}

									break
								}
							}
						} else {

							var index int
							for i := 0; i < len(reactChannelJoinMap[key].RoleEmoji[p][role]); i++ {
								if reactChannelJoinMap[key].RoleEmoji[p][role][i] == emote {

									index = i
								}
							}

							if emojiRegex != nil || emojiRegexTwo != nil {

								var emojiName []string

								// Fetches emoji API name
								re = regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
								if emojiRegex != nil {

									emojiName = re.FindAllString(emojiRegex[0], 1)

								} else if emojiRegexTwo != nil {

									emojiName = re.FindAllString(emojiRegexTwo[0], 1)
								}

								// Checks if the emoji in the map is the same as the one in the command
								if emojiName[0] == emote {

									if len(reactChannelJoinMap[commandStrings[1]].RoleEmoji[p][role]) > 1 {

										// Removes the emoji from the map
										a := reactChannelJoinMap[commandStrings[1]].RoleEmoji[p][role]
										a = append(a[:index], a[index+1:]...)
										reactChannelJoinMap[commandStrings[1]].RoleEmoji[p][role] = a

									} else {

										// Removes the entire role from the map
										delete(reactChannelJoinMap[key].RoleEmoji[p], role)
									}

									// Writes to storage
									ReactChannelJoinWrite(reactChannelJoinMap)

									// Prints success message
									_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed Emoji react join from message.")
									if err != nil {

										fmt.Println("Error:", err)
									}

									break
								} else {

									// Prints error message
									_, err = s.ChannelMessageSend(m.ChannelID, "Error: No such emoji for that messageID")
									if err != nil {

										fmt.Println("Error:", err)
									}

									break
								}

								// If it's an unicode message it removes it with this
							} else {

								// Checks if the emoji in the map is the same as the one in the command
								if commandStrings[2] == emote {

									if len(reactChannelJoinMap[commandStrings[1]].RoleEmoji[p][role]) > 1 {

										// Removes the emoji from the map
										a := reactChannelJoinMap[commandStrings[1]].RoleEmoji[p][role]
										a = append(a[:index], a[index+1:]...)
										reactChannelJoinMap[commandStrings[1]].RoleEmoji[p][role] = a

									} else {

										// Removes the entire role from the map
										delete(reactChannelJoinMap[key].RoleEmoji[p], role)
									}

									// Writes to storage
									ReactChannelJoinWrite(reactChannelJoinMap)

									_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed Emoji react join from message.")
									if err != nil {

										fmt.Println("Error:", err)
									}

									break
								} else {

									_, err = s.ChannelMessageSend(m.ChannelID, "Error: No such emoji for that messageID")
									if err != nil {

										fmt.Println("Error:", err)
									}

									break
								}
							}
						}
					}
				}
			}

			break
		}
	}
}

// Prints all currently set React Joins in memory
func viewReactJoinsCommand(s *discordgo.Session, m *discordgo.Message) {

	var line string

	if len(reactChannelJoinMap) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set react joins.")
		if err != nil {

			fmt.Println("Error:", err)
		}
	}

	// Iterates through all of the set channel joins and assigns them to a string
	for _, value := range reactChannelJoinMap {

		line = "——————\n`MessageID: " + (value.MessageID + "`\n")

		for i := 0; i < len(value.RoleEmoji); i++ {
			for role, emoji := range value.RoleEmoji[i] {

				line = line + "`" + role + "` — "
				for j := 0; j < len(emoji); j++ {

					if j != len(emoji)-1 {

						line = line + emoji[j] + ", "
					} else {

						line = line + emoji[j] + "\n"
					}
				}
			}
		}

		_, err := s.ChannelMessageSend(m.ChannelID, line)
		if err != nil {

			fmt.Println("Error:", err)
		}
	}
}

// Reads set message react join info from reactChannelJoin.json
func ReactInfoRead() {

	// Reads all the set react joins from the reactChannelJoin.json file and puts them in reactChannelJoinMap as bytes
	reactChannelJoinByte, err := ioutil.ReadFile("database/reactChannelJoin.json")
	if err != nil {

		fmt.Println(err)
	}

	MapMutex.Lock()

	// Takes all the set react join from reactChannelJoin.json from byte and puts them into the reactChannelJoinMap map
	err = json.Unmarshal(reactChannelJoinByte, &reactChannelJoinMap)
	if err != nil {

		fmt.Println(err)
	}

	MapMutex.Unlock()
}

// Saves the react channel join and parses if it already exists
func SaveReactJoin(messageID string, role string, emoji string) {

	var (
		temp        ReactChannelJoinStruct
		emojiExists = false
	)

	if reactChannelJoinMap[messageID] != nil {

		temp = *reactChannelJoinMap[messageID]

		MapMutex.Lock()

		// Sets MessageID
		temp.MessageID = messageID

		if temp.RoleEmoji == nil {

			temp.RoleEmoji = append(temp.RoleEmoji, EmojiRoleMap)
		}

		for i := 0; i < len(temp.RoleEmoji); i++ {
			if temp.RoleEmoji[i][role] == nil {

				temp.RoleEmoji[i][role] = append(temp.RoleEmoji[i][role], emoji)
			}

			for j := 0; j < len(temp.RoleEmoji[i][role]); j++ {
				if temp.RoleEmoji[i][role][j] == emoji {

					emojiExists = true
				}
			}
			if emojiExists == false {

				temp.RoleEmoji[i][role] = append(temp.RoleEmoji[i][role], emoji)
			}
		}

		reactChannelJoinMap[messageID] = &temp

		MapMutex.Unlock()

	} else {

		MapMutex.Lock()

		// Sets messageID
		temp.MessageID = messageID

		// Initializes temp.RoleEmoji if it's nil
		EmojiRoleMapDummy := make(map[string][]string)
		if temp.RoleEmoji == nil {

			temp.RoleEmoji = append(temp.RoleEmoji, EmojiRoleMapDummy)
		}

		for i := 0; i < len(temp.RoleEmoji); i++ {
			if temp.RoleEmoji[i][role] == nil {

				temp.RoleEmoji[i][role] = append(temp.RoleEmoji[i][role], emoji)
			}
		}

		reactChannelJoinMap[messageID] = &temp

		MapMutex.Unlock()
	}
}

// Writes react channel join info to ReactChannelJoinWrite.json
func ReactChannelJoinWrite(info map[string]*ReactChannelJoinStruct) {

	MapMutex.Lock()

	// Turns info slice into byte ready to be pushed to file
	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	MapMutex.Unlock()

	// Writes to file
	err = ioutil.WriteFile("database/reactChannelJoin.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

func init() {
	add(&command{
		execute:  setReactJoinCommand,
		trigger:  "setreactjoin",
		aliases:  []string{"setreact"},
		desc:     "Sets a react join on a specific message, role and emote.",
		elevated: true,
	})
	add(&command{
		execute:  removeReactJoinCommand,
		trigger:  "removereactjoin",
		aliases:  []string{"removereact"},
		desc:     "Removes a set react join.",
		elevated: true,
	})
	add(&command{
		execute:  viewReactJoinsCommand,
		trigger:  "viewreactjoins",
		aliases:  []string{"viewreacts", "viewreact"},
		desc:     "Views all set react joins.",
		elevated: true,
	})
}