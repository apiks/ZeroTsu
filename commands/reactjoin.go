package commands

import (
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
	reactChannelJoinMap = make(map[string]*reactChannelJoinStruct)
	EmojiRoleMap        = make(map[string][]string)
)

type reactChannelJoinStruct struct {
	RoleEmojiMap []map[string][]string `json:"roleEmoji"`
}

// Gives a specific role to a user if they react
func ReactJoinHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				return
			}
		}
	}()

	// Checks if a react channel join is set for that specific message and emoji and continues if true
	misc.MapMutex.Lock()
	if reactChannelJoinMap[r.MessageID] == nil {
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Pulls all of the server roles
	roles, err := s.GuildRoles(config.ServerID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Puts the react API emoji name to lowercase so it is valid with the storage emoji name
	reactLowercase := strings.ToLower(r.Emoji.APIName())

	misc.MapMutex.Lock()
	for _, roleEmojiMap := range reactChannelJoinMap[r.MessageID].RoleEmojiMap {
		for role, emojiSlice := range roleEmojiMap {
			for _, emoji := range emojiSlice {
				if reactLowercase != emoji {
					continue
				}

				// If the role is over 17 in characters it checks if it's a valid role ID and gives the role if so
				// Otherwise it iterates through all roles to find the proper one
				if len(role) >= 17 {
					if _, err := strconv.ParseInt(role, 10, 64); err == nil {
						// Gives the role
						err := s.GuildMemberRoleAdd(config.ServerID, r.UserID, role)
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
							if err != nil {
								misc.MapMutex.Unlock()
								return
							}
							misc.MapMutex.Unlock()
							return
						}
						misc.MapMutex.Unlock()
						return
					}
				}
				for _, serverRole := range roles {
					if serverRole.Name == role {
						// Gives the role
						err := s.GuildMemberRoleAdd(config.ServerID, r.UserID, serverRole.ID)
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
							if err != nil {
								misc.MapMutex.Unlock()
								return
							}
							misc.MapMutex.Unlock()
							return
						}
					}
				}
			}
		}
	}
	misc.MapMutex.Unlock()
}

// Removes a role from user if they unreact
func ReactRemoveHandler(s *discordgo.Session, r *discordgo.MessageReactionRemove) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				return
			}
		}
	}()

	// Checks if a react channel join is set for that specific message and emoji and continues if true
	misc.MapMutex.Lock()
	if reactChannelJoinMap[r.MessageID] == nil {
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Pulls all of the server roles
	roles, err := s.GuildRoles(config.ServerID)
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Puts the react API emoji name to lowercase so it is valid with the storage emoji name
	reactLowercase := strings.ToLower(r.Emoji.APIName())

	misc.MapMutex.Lock()
	for _, roleEmojiMap := range reactChannelJoinMap[r.MessageID].RoleEmojiMap {
		for role, emojiSlice := range roleEmojiMap {
			for _, emoji := range emojiSlice {
				if reactLowercase != emoji {
					continue
				}

				// If the role is over 17 in characters it checks if it's a valid role ID and removes the role if so
				// Otherwise it iterates through all roles to find the proper one
				if len(role) >= 17 {
					if _, err := strconv.ParseInt(role, 10, 64); err == nil {
						// Removes the role
						err := s.GuildMemberRoleRemove(config.ServerID, r.UserID, role)
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
							if err != nil {
								misc.MapMutex.Unlock()
								return
							}
							misc.MapMutex.Unlock()
							return
						}
						misc.MapMutex.Unlock()
						return
					}
				}
				for _, serverRole := range roles {
					if serverRole.Name == role {
						// Removes the role
						err := s.GuildMemberRoleRemove(config.ServerID, r.UserID, serverRole.ID)
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
							if err != nil {
								misc.MapMutex.Unlock()
								return
							}
							misc.MapMutex.Unlock()
							return
						}
					}
				}
			}
		}
	}
	misc.MapMutex.Unlock()
}

// Sets react joins per specific message and emote
func setReactJoinCommand (s *discordgo.Session, m *discordgo.Message) {

	var roleExists bool

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 4)

	if len(commandStrings) != 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"setreact [messageID] [emoji] [role]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Checks if it's a valid messageID
	num, err := strconv.Atoi(commandStrings[1])
	if err != nil || num < 17 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid messageID.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fetches all server roles
	roles, err := s.GuildRoles(config.ServerID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Checks if the role exists in the server roles
	for _, role := range roles {
		if strings.ToLower(role.Name) == commandStrings[3] {
			roleExists = true
			break
		}
	}
	if !roleExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid role.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Parses if it's custom emoji or unicode emoji
	re := regexp.MustCompile("(?i)<:+([a-zA-Z]|[0-9])+:+[0-9]+>")
	emojiRegex := re.FindAllString(messageLowercase, 1)
	if emojiRegex != nil {

		// Fetches emoji API name
		re = regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
		emojiName := re.FindAllString(emojiRegex[0], 1)

		// Sets the data in memory to be ready for writing
		SaveReactJoin(commandStrings[1], commandStrings[3], emojiName[0])

		// Writes the data to storage
		ReactChannelJoinWrite(reactChannelJoinMap)

		// Reacts with the set emote if possible and gives success
		_ = s.MessageReactionAdd(m.ChannelID, commandStrings[1], emojiName[0])
		_, err = s.ChannelMessageSend(m.ChannelID, "Success! React channel join set.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {
				return
			}
			return
		}
		return
	}

	// If the above is false, it's a non-valid emoji or an unicode emoji (the latter preferably) and saves that

	// Sets the data in memory to be ready for writing
	SaveReactJoin(commandStrings[1], commandStrings[3], commandStrings[2])

	// Writes the data to storage
	ReactChannelJoinWrite(reactChannelJoinMap)

	// Reacts with the set emote if possible
	_ = s.MessageReactionAdd(m.ChannelID, commandStrings[1], commandStrings[2])
	_, err = s.ChannelMessageSend(m.ChannelID, "Success! React channel join set.")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func removeReactJoinCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		messageExists bool
		validEmoji =  false

		messageID     string
		emojiRegexAPI []string
		emojiAPI	  []string
	)

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 3)

	if len(commandStrings) != 3 && len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+config.BotPrefix+"removereact [messageID] Optional[emoji]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Checks if it's a valid messageID
	num, err := strconv.Atoi(commandStrings[1])
	if err != nil || num < 17 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid messageID.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	misc.MapMutex.Lock()
	if len(reactChannelJoinMap) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set react joins.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}
	// Checks if the messageID already exists in the map
	for k := range reactChannelJoinMap {
		if commandStrings[1] == k {
			messageExists = true
			messageID = k
			break
		}
	}
	misc.MapMutex.Unlock()
	if messageExists == false {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: No such messageID is set in storage")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Removes the entire message from the map and writes to storage
	misc.MapMutex.Lock()
	if len(commandStrings) == 2 {
		delete(reactChannelJoinMap, commandStrings[1])
		ReactChannelJoinWrite(reactChannelJoinMap)
		_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed entire message emoji react join.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	if reactChannelJoinMap[messageID].RoleEmojiMap == nil {
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Parses if it's custom emoji or unicode
	re := regexp.MustCompile("(?i)<:+([a-zA-Z]|[0-9])+:+[0-9]+>")
	emojiRegex := re.FindAllString(commandStrings[2], 1)
	if emojiRegex == nil {
		// Second parser if it's custom emoji or unicode but for emoji API name instead
		reAPI := regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
		emojiRegexAPI = reAPI.FindAllString(commandStrings[2], 1)
	}

	misc.MapMutex.Lock()
	for storageMessageID := range reactChannelJoinMap[messageID].RoleEmojiMap {
		for role, emojiSlice := range reactChannelJoinMap[messageID].RoleEmojiMap[storageMessageID] {
			for index, emoji := range emojiSlice {

				// Checks for unicode emoji
				if len(emojiRegex) == 0 && len(emojiRegexAPI) == 0 {
					if commandStrings[2] == emoji {
						validEmoji = true
					}
					// Checks for non-unicode emoji
				} else {
					// Trims non-unicode emoji name to fit API emoji name
					re = regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
					if len(emojiRegex) == 0 {
						if len(emojiRegexAPI) != 0 {
							emojiAPI = re.FindAllString(emojiRegexAPI[0], 1)
							if emoji == emojiAPI[0] {
								validEmoji = true
							}
						}
					} else {
						emojiAPI = re.FindAllString(emojiRegex[0], 1)
						if emoji == emojiAPI[0] {
							validEmoji = true
						}
					}
				}

				// Delete only if it's a valid emoji in map
				if validEmoji {
					// Delete the entire message from map if it's the only set emoji react join
					if len(reactChannelJoinMap[messageID].RoleEmojiMap[storageMessageID]) == 1 && len(reactChannelJoinMap[messageID].RoleEmojiMap[storageMessageID][role]) == 1 {
						delete(reactChannelJoinMap, commandStrings[1])
						ReactChannelJoinWrite(reactChannelJoinMap)
						_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed emoji react join from message.")
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
							if err != nil {
								misc.MapMutex.Unlock()
								return
							}
							misc.MapMutex.Unlock()
							return
						}
						// Delete only the role from map if other set react join roles exist in the map
					} else if len(reactChannelJoinMap[messageID].RoleEmojiMap[storageMessageID][role]) == 1 {
						delete(reactChannelJoinMap[messageID].RoleEmojiMap[storageMessageID], role)
						ReactChannelJoinWrite(reactChannelJoinMap)
						_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed emoji react join from message.")
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
							if err != nil {
								misc.MapMutex.Unlock()
								return
							}
							misc.MapMutex.Unlock()
							return
						}
						// Delete only that specific emoji for that specific role
					} else {
						a := reactChannelJoinMap[commandStrings[1]].RoleEmojiMap[storageMessageID][role]
						a = append(a[:index], a[index+1:]...)
						reactChannelJoinMap[commandStrings[1]].RoleEmojiMap[storageMessageID][role] = a
						_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed emoji react join from message.")
						if err != nil {
							_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
							if err != nil {
								misc.MapMutex.Unlock()
								return
							}
							misc.MapMutex.Unlock()
							return
						}
					}
					misc.MapMutex.Unlock()
					return
				}

			}
		}
	}
	misc.MapMutex.Unlock()

	// If it comes this far it means it's an invalid emoji
	if emojiRegex == nil && emojiRegexAPI == nil {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Invalid emoji. Please input a valid emoji or emoji API name.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}
}

// Prints all currently set React Joins in memory
func viewReactJoinsCommand(s *discordgo.Session, m *discordgo.Message) {

	var line string

	misc.MapMutex.Lock()
	if len(reactChannelJoinMap) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set react joins.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	// Iterates through all of the set channel joins and assigns them to a string
	for messageID, value := range reactChannelJoinMap {

		// Formats message
		line = "——————\n`MessageID: " + (messageID + "`\n")
		for i := 0; i < len(value.RoleEmojiMap); i++ {
			for role, emoji := range value.RoleEmojiMap[i] {
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
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
	}
	misc.MapMutex.Unlock()
}

// Reads set message react join info from reactChannelJoin.json
func ReactInfoRead() {

	// Reads all the set react joins from the reactChannelJoin.json file and puts them in reactChannelJoinMap as bytes
	reactChannelJoinByte, err := ioutil.ReadFile("database/reactChannelJoin.json")
	if err != nil {
		return
	}

	// Takes all the set react join from reactChannelJoin.json from byte and puts them into the reactChannelJoinMap map
	misc.MapMutex.Lock()
	err = json.Unmarshal(reactChannelJoinByte, &reactChannelJoinMap)
	if err != nil {
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()
}

// Writes react channel join info to ReactChannelJoinWrite.json
func ReactChannelJoinWrite(info map[string]*reactChannelJoinStruct) {

	// Turns info slice into byte ready to be pushed to file
	marshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		return
	}

	// Writes to file
	err = ioutil.WriteFile("database/reactChannelJoin.json", marshaledStruct, 0644)
	if err != nil {
		return
	}
}

// Saves the react channel join and parses if it already exists
func SaveReactJoin(messageID string, role string, emoji string) {

	var (
		temp		  reactChannelJoinStruct
		emojiExists = false
	)

	// Uses this if the message already has a set emoji react
	misc.MapMutex.Lock()
	if reactChannelJoinMap[messageID] != nil {
		temp = *reactChannelJoinMap[messageID]

		if temp.RoleEmojiMap == nil {
			temp.RoleEmojiMap = append(temp.RoleEmojiMap, EmojiRoleMap)
		}

		for i := 0; i < len(temp.RoleEmojiMap); i++ {
			if temp.RoleEmojiMap[i][role] == nil {
				temp.RoleEmojiMap[i][role] = append(temp.RoleEmojiMap[i][role], emoji)
			}

			for j := 0; j < len(temp.RoleEmojiMap[i][role]); j++ {
				if temp.RoleEmojiMap[i][role][j] == emoji {
					emojiExists = true
					break
				}
			}
			if !emojiExists {
				temp.RoleEmojiMap[i][role] = append(temp.RoleEmojiMap[i][role], emoji)
			}
		}

		reactChannelJoinMap[messageID] = &temp
		misc.MapMutex.Unlock()
		return
	}

	// Initializes temp.RoleEmoji if the message doesn't have a set emoji react
	EmojiRoleMapDummy := make(map[string][]string)
	if temp.RoleEmojiMap == nil {
		temp.RoleEmojiMap = append(temp.RoleEmojiMap, EmojiRoleMapDummy)
	}

	for i := 0; i < len(temp.RoleEmojiMap); i++ {
		if temp.RoleEmojiMap[i][role] == nil {
			temp.RoleEmojiMap[i][role] = append(temp.RoleEmojiMap[i][role], emoji)
		}
	}

	reactChannelJoinMap[messageID] = &temp
	misc.MapMutex.Unlock()
}

func init() {
	add(&command{
		execute:  setReactJoinCommand,
		trigger:  "setreactjoin",
		aliases:  []string{"setreact"},
		desc:     "Sets a react join on a specific message, role and emote.",
		elevated: true,
		category: "reacts",
	})
	add(&command{
		execute:  removeReactJoinCommand,
		trigger:  "removereactjoin",
		aliases:  []string{"removereact"},
		desc:     "Removes a set react join.",
		elevated: true,
		category: "reacts",
	})
	add(&command{
		execute:  viewReactJoinsCommand,
		trigger:  "viewreactjoins",
		aliases:  []string{"viewreacts", "viewreact", "reacts"},
		desc:     "Views all set react joins.",
		elevated: true,
		category: "reacts",
	})
}