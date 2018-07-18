package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/misc"
	"regexp"
	"strings"

	"encoding/json"
	"github.com/r-anime/ZeroTsu/config"
	"io/ioutil"
	"strconv"
	"sync"
)

var (
	reactChannelJoinMap = make(map[string]*ReactChannelJoinStruct)
	EmojiRoleMap        = make(map[string][]string)
	MapMutex            sync.Mutex
)

// Creates a new struct to keep the parameters passed in storage
type ReactChannelJoinStruct struct {
	MessageID string                `json:"messageID"`
	RoleEmoji []map[string][]string `json:"roleEmoji"`
}

// Reads set message react join info from reactChannelJoin.json
func setReactInfoRead() {

	MapMutex.Lock()

	// Reads all the set react joins from the reactChannelJoin.json file and puts them in reactChannelJoinMap as bytes
	reactChannelJoinByte, err := ioutil.ReadFile("database/reactChannelJoin.json")
	if err != nil {

		fmt.Println(err)
	}

	// Takes all the set react join from reactChannelJoin.json from byte and puts them into the reactChannelJoinMap map
	err = json.Unmarshal(reactChannelJoinByte, &reactChannelJoinMap)
	if err != nil {

		fmt.Println(err)
	}

	MapMutex.Unlock()
}

// Saves the react channel join and parses if it already exists so multiple can be kept
func SaveReactJoin(messageID string, role string, emoji string) {

	var (
		temp        ReactChannelJoinStruct
		emojiExists = false
	)

	if reactChannelJoinMap[messageID] != nil {

		temp = *reactChannelJoinMap[messageID]

		MapMutex.Lock()

		//Sets MessageID
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

		//Sets messageID
		temp.MessageID = messageID

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

	//Turns info slice into byte ready to be pushed to file
	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	//Writes to file
	err = ioutil.WriteFile("database/reactChannelJoin.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}

	MapMutex.Unlock()
}

// Allows mods to set react joins per specific message and emote
func SetReactChannelHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

			//Pulls info on message author
			mem, err := s.State.Member(config.ServerID, m.Author.ID)
			if err != nil {
				mem, err = s.GuildMember(config.ServerID, m.Author.ID)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}

			//Checks for mod permissions
			if misc.HasPermissions(mem) {
				if m.Author.ID == config.BotID {
					return
				}

				//Puts the command to lowercase
				messageLowercase := strings.ToLower(m.Content)

				//Checks if BotPrefix + setreactjoin was used
				if strings.HasPrefix(messageLowercase, config.BotPrefix+"setreactjoin ") && (messageLowercase != (config.BotPrefix + "setreactjoin")) {

					// Splits the message parameters via space
					messageSplit := strings.Split(messageLowercase, " ")

					// Checks if the message has the necessary parameters
					if len(messageSplit) == 4 && len(messageSplit[1]) == 18 {

						// Checks if the role exists on the server
						var roleExists = false
						roles, err := s.GuildRoles(config.ServerID)
						if err != nil {

							fmt.Println("Error: ", err)
						}

						// Iterates through all of the roles and sets bool value to true if the role exists
						for j := 0; j < len(roles); j++ {

							role := strings.ToLower(roles[j].Name)

							if role == messageSplit[2] {
								roleExists = true
								break
							}
						}

						// Continues down if the role is a valid one
						if roleExists == true {

							// Parses if it's custom emoji or unicode
							re := regexp.MustCompile("(?i)<:+([a-zA-Z]|[0-9])+:+[0-9]+>")
							emojiRegex := re.FindAllString(messageLowercase, 1)

							if emojiRegex != nil {

								// Fetches emoji API name
								re := regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
								emojiName := re.FindAllString(emojiRegex[0], 1)

								// Reads the data from storage
								setReactInfoRead()

								// Sets the data in memory to be ready for writing
								SaveReactJoin(messageSplit[1], messageSplit[2], emojiName[0])

								// Writes the data to storage
								ReactChannelJoinWrite(reactChannelJoinMap)

								// Reacts with the set emote
								err := s.MessageReactionAdd(m.ChannelID, messageSplit[1], emojiName[0])
								if err != nil {

									// Sets and prints error message
									message, err := s.ChannelMessage(m.ChannelID, messageSplit[1])
									if err != nil {
									} else if message.Reactions != nil {
										if message.Reactions[0].Count == 20 {

											// Prints error message
											_, err = s.ChannelMessageSend(m.ChannelID, "Error: Reached the max reaction limit. (20)")
											if err != nil {

												fmt.Println("Error: ", err)
											}
										}
									}
								}

								// Prints success message
								_, err = s.ChannelMessageSend(m.ChannelID, "Success! React channel join set.")
								if err != nil {

									fmt.Println("Error: ", err)
								}

								// Else it's a non-valid emoji or an unicode emoji (the latter preferably) and saves that
							} else {

								// Sets the data in memory to be ready for writing
								SaveReactJoin(messageSplit[1], messageSplit[2], messageSplit[3])

								// Writes the data to storage
								ReactChannelJoinWrite(reactChannelJoinMap)

								// Reacts with the set emote if it's in the channel
								err := s.MessageReactionAdd(m.ChannelID, messageSplit[1], messageSplit[3])
								if err != nil {

									message, err := s.ChannelMessage(m.ChannelID, messageSplit[1])
									if err != nil {
									} else if message != nil {
										if len(message.Reactions) > 0 {
											if message.Reactions[0] != nil {
												if message.Reactions[0].Count == 20 {

													// Prints error message
													_, err = s.ChannelMessageSend(m.ChannelID, "Error: Reached the max reaction limit. (20)")
													if err != nil {

														fmt.Println("Error: ", err)
													}
												}
											}
										}
									}
								}

								// Prints success message
								_, err = s.ChannelMessageSend(m.ChannelID, "Success! React channel join set.")
								if err != nil {

									fmt.Println("Error: ", err)
								}
							}

						} else {

							// Prints error message
							_, err = s.ChannelMessageSend(m.ChannelID, "Error: Role doesn't exist.")
							if err != nil {

								fmt.Println("Error: ", err)
							}
						}
					} else if len(messageSplit) > 4 {

						// Prints error message
						_, err = s.ChannelMessageSend(m.ChannelID, "Error: Too many parameters. Use `"+config.BotPrefix+"setreactjoin [messageID] [role] [emoji] `")
						if err != nil {

							fmt.Println("Error: ", err)
						}
					} else if len(messageSplit) < 4 {

						// Prints error message
						_, err = s.ChannelMessageSend(m.ChannelID, "Error: Too little parameters. Use `"+config.BotPrefix+"setreactjoin [messageID] [role] [emoji] `")
						if err != nil {

							fmt.Println("Error: ", err)
						}
					}
				}
			}
		}
	}
}

// Allows mods to view the set react joins per specific message and emote
func ViewSetReactJoinsHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

			//Pulls info on message author
			mem, err := s.State.Member(config.ServerID, m.Author.ID)
			if err != nil {
				mem, err = s.GuildMember(config.ServerID, m.Author.ID)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}

			//Checks for mod permissions
			if misc.HasPermissions(mem) {
				if m.Author.ID == config.BotID {
					return
				}

				//Puts the command to lowercase
				messageLowercase := strings.ToLower(m.Content)

				//Checks if BotPrefix + viewreacts was used
				if strings.HasPrefix(messageLowercase, config.BotPrefix+"viewreacts") {

					// Reads all the set react joins from storage
					setReactInfoRead()

					// Checks if the map is empty, gives error if it is
					if len(reactChannelJoinMap) != 0 {

						// Variable in which we'll keep all of the values
						var line string

						// Iterates through all of the set channel joins and assigns them to a string
						for _, value := range reactChannelJoinMap {

							//line = ""

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

							// Sends the message
							_, err = s.ChannelMessageSend(m.ChannelID, line)
							if err != nil {

								fmt.Println("Error: ", err)
							}
						}
					} else {

						_, err = s.ChannelMessageSend(m.ChannelID, "Error: There are no set react joins.")
						if err != nil {

							fmt.Println("Error: ", err)
						}
					}
				}
			}
		}
	}
}

// Allows mods to remove the set react joins per specific message or emote
func RemoveReactJoinHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

			//Pulls info on message author
			mem, err := s.State.Member(config.ServerID, m.Author.ID)
			if err != nil {
				mem, err = s.GuildMember(config.ServerID, m.Author.ID)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}

			//Checks for mod permissions
			if misc.HasPermissions(mem) {
				if m.Author.ID == config.BotID {
					return
				}

				//Puts the command to lowercase
				messageLowercase := strings.ToLower(m.Content)

				//Checks if BotPrefix + removereactjoin was used
				if strings.HasPrefix(messageLowercase, config.BotPrefix+"removereactjoin ") && (messageLowercase != (config.BotPrefix + "removereactjoin")) {

					// Splits the message parameters via space
					messageSplit := strings.Split(messageLowercase, " ")

					// Reads all the set react joins from storage
					setReactInfoRead()

					// Checks if the map is empty and displays error message if true
					if len(reactChannelJoinMap) != 0 {

						// Checks if it's a messageID for entire message removal and removes it if true. Else it removes emojis only. Also checks if messageID exists in map
						if len(messageSplit[1]) == 18 && len(messageSplit) == 2 {
							if _, err := strconv.ParseInt(messageSplit[1], 10, 64); err == nil {

								// Checks if the messageID already exists in the map, if not it displays error
								var messageExists = false
								for k := range reactChannelJoinMap {

									if messageSplit[1] == k {

										messageExists = true
									}
								}

								if messageExists == true {
									// Removes the entire message from the map
									delete(reactChannelJoinMap, messageSplit[1])
									ReactChannelJoinWrite(reactChannelJoinMap)

									// Prints success message
									_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed entire message emoji react join.")
									if err != nil {

										fmt.Println("Error: ", err)
									}
								} else {

									// Prints error message
									_, err = s.ChannelMessageSend(m.ChannelID, "Error: No such messageID is set in storage")
								}
							}

							// Parses if message is correct
						} else if len(messageSplit[1]) == 18 && len(messageSplit) > 2 {

							// Checks if the messageID already exists in the map, if not it displays error
							var messageExists = false
							var key string
							for k := range reactChannelJoinMap {

								if messageSplit[1] == k {

									messageExists = true
									key = k
								}
							}

							if messageExists == true {

								if reactChannelJoinMap[key].RoleEmoji != nil {

									for p := 0; p < len(reactChannelJoinMap[key].RoleEmoji); p++ {
										for role := range reactChannelJoinMap[key].RoleEmoji[p] {
											for _, emote := range reactChannelJoinMap[key].RoleEmoji[p][role] {

												// Parses if it's custom emoji or unicode
												re := regexp.MustCompile("(?i)<:+([a-zA-Z]|[0-9])+:+[0-9]+>")
												emojiRegex := re.FindAllString(messageSplit[2], 1)

												// Second parser if it's custom emoji or unicode but for emoji API name instead
												reTwo := regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
												emojiRegexTwo := reTwo.FindAllString(messageSplit[2], 1)

												// Check whether emojiRegexTwo is empty or not to avoid index error
												var regexTwoBool = false
												if len(emojiRegexTwo) > 0 {
													if emote == emojiRegexTwo[0] {

														regexTwoBool = true
													}
												}

												// Formats the messageSplit[2] unicode properly
												unicode := messageSplit[2]

												// Checks if either regexTwoBool exists or the emote <-> unicode are the same
												if emote == unicode || regexTwoBool == true {

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

																// Removes the entire message from the map
																delete(reactChannelJoinMap, messageSplit[1])
																ReactChannelJoinWrite(reactChannelJoinMap)

																// Prints success message
																_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed Emoji react join from message.")
																if err != nil {

																	fmt.Println("Error: ", err)
																}

																break
															} else {

																// Prints error message
																_, err = s.ChannelMessageSend(m.ChannelID, "Error: No such emoji for that messageID")
																if err != nil {

																	fmt.Println("Error: ", err)
																}
															}

															break

															// If it's an unicode message it removes it with this
														} else {

															// Checks if the emoji in the map is the same as the one in the command
															if messageSplit[2] == emote {

																// Removes the entire message from the map
																delete(reactChannelJoinMap, messageSplit[1])
																ReactChannelJoinWrite(reactChannelJoinMap)

																// Prints success message
																_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed Emoji react join from message.")
																if err != nil {

																	fmt.Println("Error: ", err)
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

																if len(reactChannelJoinMap[messageSplit[1]].RoleEmoji[p][role]) > 1 {

																	// Removes the emoji from the map
																	a := reactChannelJoinMap[messageSplit[1]].RoleEmoji[p][role]
																	a = append(a[:index], a[index+1:]...)
																	reactChannelJoinMap[messageSplit[1]].RoleEmoji[p][role] = a

																} else {

																	// Removes the entire role from the map
																	delete(reactChannelJoinMap[key].RoleEmoji[p], role)
																}

																// Writes to storage
																ReactChannelJoinWrite(reactChannelJoinMap)

																// Prints success message
																_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed Emoji react join from message.")
																if err != nil {

																	fmt.Println("Error: ", err)
																}

																break
															} else {

																// Prints error message
																_, err = s.ChannelMessageSend(m.ChannelID, "Error: No such emoji for that messageID")
																if err != nil {

																	fmt.Println("Error: ", err)
																}

																break
															}

															// If it's an unicode message it removes it with this
														} else {

															// Checks if the emoji in the map is the same as the one in the command
															if messageSplit[2] == emote {

																if len(reactChannelJoinMap[messageSplit[1]].RoleEmoji[p][role]) > 1 {

																	// Removes the emoji from the map
																	a := reactChannelJoinMap[messageSplit[1]].RoleEmoji[p][role]
																	a = append(a[:index], a[index+1:]...)
																	reactChannelJoinMap[messageSplit[1]].RoleEmoji[p][role] = a

																} else {

																	// Removes the entire role from the map
																	delete(reactChannelJoinMap[key].RoleEmoji[p], role)
																}

																// Writes to storage
																ReactChannelJoinWrite(reactChannelJoinMap)

																// Prints success message
																_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed Emoji react join from message.")
																if err != nil {

																	fmt.Println("Error: ", err)
																}

																break
															} else {

																// Prints error message
																_, err = s.ChannelMessageSend(m.ChannelID, "Error: No such emoji for that messageID")
																if err != nil {

																	fmt.Println("Error: ", err)
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

							} else if len(messageSplit[1]) != 18 {

								// Prints error message
								_, err = s.ChannelMessageSend(m.ChannelID, "Error: Wrong messageID value.")
								if err != nil {

									fmt.Println("Error: ", err)
								}
							}

							// If there are no set emoji react joins print error message
						} else {

							// Prints error message
							_, err = s.ChannelMessageSend(m.ChannelID, "Error: Wrong command. Please use `"+config.BotPrefix+"removereactjoin [messageID] OPTIONAL[emoji]`")
						}
					} else {

						// Prints error message
						_, err = s.ChannelMessageSend(m.ChannelID, "Error: There are no set react channel joins")
					}
				}
			}
		}
	}
}

// Adds role to user if they react
func ReactJoinHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	// Reads all the set react joins from storage
	setReactInfoRead()

	// Checks if a react channel join is set for that specific message and emoji and continues if true
	if reactChannelJoinMap[r.MessageID] != nil {

		// Puts the react API name to lowercase so it is valid with the storage emoji name
		reactLowercase := strings.ToLower(r.Emoji.APIName())

		// Checks if it's the correct message and emoji before going down
		if reactChannelJoinMap[r.MessageID].MessageID == r.MessageID {

			for p := 0; p < len(reactChannelJoinMap[r.MessageID].RoleEmoji); p++ {
				for role := range reactChannelJoinMap[r.MessageID].RoleEmoji[p] {
					for _, emote := range reactChannelJoinMap[r.MessageID].RoleEmoji[p][role] {
						if reactLowercase == emote {

							if len(role) == 18 {
								if _, err := strconv.ParseInt(role, 10, 64); err == nil {

									// Gives the role
									err := s.GuildMemberRoleAdd(config.ServerID, r.UserID, role)
									if err != nil {

										fmt.Println("Error: ", err)
									}
								}
							} else {

								// Pulls all of the server roles
								roles, err := s.GuildRoles(config.ServerID)
								if err != nil {

									fmt.Println("Error: ", err)
								}

								// Iterates through all of the server roles and gives the role to the user if the set role exists
								for i := 0; i < len(roles); i++ {
									if roles[i].Name == role {

										// Gives the role
										err := s.GuildMemberRoleAdd(config.ServerID, r.UserID, roles[i].ID)
										if err != nil {

											fmt.Println("Error: ", err)
										}
									}
								}
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

	// Reads all the set react joins from storage
	setReactInfoRead()

	// Checks if a react channel join is set for that specific message and emoji and continues if true
	if reactChannelJoinMap[r.MessageID] != nil {

		// Puts the react API name to lowercase so it is valid with the storage emoji name
		reactLowercase := strings.ToLower(r.Emoji.APIName())

		// Checks if it's the correct message and emoji before going down
		if reactChannelJoinMap[r.MessageID].MessageID == r.MessageID {

			for p := 0; p < len(reactChannelJoinMap[r.MessageID].RoleEmoji); p++ {
				for role := range reactChannelJoinMap[r.MessageID].RoleEmoji[p] {
					for _, emote := range reactChannelJoinMap[r.MessageID].RoleEmoji[p][role] {
						if reactLowercase == emote {

							if len(role) == 18 {
								if _, err := strconv.ParseInt(role, 10, 64); err == nil {

									// Removes the role
									err := s.GuildMemberRoleRemove(config.ServerID, r.UserID, role)
									if err != nil {

										fmt.Println("Error: ", err)
									}
								}
							} else {

								// Pulls all of the server roles
								roles, err := s.GuildRoles(config.ServerID)
								if err != nil {

									fmt.Println("Error: ", err)
								}

								// Iterates through all of the server roles and removes the role from the user if the set role exists
								for i := 0; i < len(roles); i++ {
									if roles[i].Name == role {

										// Removes the role
										err := s.GuildMemberRoleRemove(config.ServerID, r.UserID, roles[i].ID)
										if err != nil {

											fmt.Println("Error: ", err)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}
