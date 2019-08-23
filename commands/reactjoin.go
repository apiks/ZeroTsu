package commands

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Gives a specific role to a user if they react
func ReactJoinHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in ReactJoinHandler")
		}
	}()

	if r.GuildID == "" {
		return
	}

	// Checks if a react channel join is set for that specific message and emoji and continues if true
	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[r.GuildID].GuildConfig.BotLog.ID

	if misc.GuildMap[r.GuildID].ReactJoinMap[r.MessageID] == nil {
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Return if the one reacting is this BOT
	if r.UserID == s.State.SessionID {
		return
	}

	// Pulls all of the server roles
	roles, err := s.GuildRoles(r.GuildID)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Puts the react API emoji name to lowercase so it is valid with the storage emoji name
	reactLowercase := strings.ToLower(r.Emoji.APIName())

	misc.MapMutex.Lock()
	for _, roleEmojiMap := range misc.GuildMap[r.GuildID].ReactJoinMap[r.MessageID].RoleEmojiMap {
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
						err := s.GuildMemberRoleAdd(r.GuildID, r.UserID, role)
						if err != nil {
							_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
					if strings.ToLower(serverRole.Name) == strings.ToLower(role) {
						// Gives the role
						err := s.GuildMemberRoleAdd(r.GuildID, r.UserID, serverRole.ID)
						if err != nil {
							_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
			log.Println(rec)
			log.Println("Recovery in ReactRemoveHandler")
		}
	}()

	if r.GuildID == "" {
		return
	}

	// Checks if a react channel join is set for that specific message and emoji and continues if true
	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[r.GuildID].GuildConfig.BotLog.ID

	if misc.GuildMap[r.GuildID].ReactJoinMap[r.MessageID] == nil {
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Return if the one unreacting is this BOT
	if r.UserID == s.State.SessionID {
		return
	}

	// Pulls all of the server roles
	roles, err := s.GuildRoles(r.GuildID)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Puts the react API emoji name to lowercase so it is valid with the storage emoji name
	reactLowercase := strings.ToLower(r.Emoji.APIName())

	misc.MapMutex.Lock()
	for _, roleEmojiMap := range misc.GuildMap[r.GuildID].ReactJoinMap[r.MessageID].RoleEmojiMap {
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
						err := s.GuildMemberRoleRemove(r.GuildID, r.UserID, role)
						if err != nil {
							_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
					if strings.ToLower(serverRole.Name) == strings.ToLower(role) {
						// Removes the role
						err := s.GuildMemberRoleRemove(r.GuildID, r.UserID, serverRole.ID)
						if err != nil {
							_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
func setReactJoinCommand(s *discordgo.Session, m *discordgo.Message) {

	var roleExists bool

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 4)

	if len(commandStrings) != 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"setreact [messageID] [emoji] [role]`")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fetches all server roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
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
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
		SaveReactJoin(commandStrings[1], commandStrings[3], emojiName[0], m.GuildID)

		// Writes the data to storage
		misc.MapMutex.Lock()
		err = misc.ReactJoinWrite(misc.GuildMap[m.GuildID].ReactJoinMap, m.GuildID)
		if err != nil {
			misc.MapMutex.Unlock()
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		misc.MapMutex.Unlock()

		// Reacts with the set emote if possible and gives success
		_ = s.MessageReactionAdd(m.ChannelID, commandStrings[1], emojiName[0])
		_, err = s.ChannelMessageSend(m.ChannelID, "Success! React channel join set.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error())
			if err != nil {
				return
			}
			return
		}
		return
	}

	// If the above is false, it's a non-valid emoji or an unicode emoji (the latter preferably) and saves that

	// Sets the data in memory to be ready for writing
	SaveReactJoin(commandStrings[1], commandStrings[3], commandStrings[2], m.GuildID)

	// Writes the data to storage
	misc.MapMutex.Lock()
	err = misc.ReactJoinWrite(misc.GuildMap[m.GuildID].ReactJoinMap, m.GuildID)
	if err != nil {
		misc.MapMutex.Unlock()
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	misc.MapMutex.Unlock()

	// Reacts with the set emote if possible
	_ = s.MessageReactionAdd(m.ChannelID, commandStrings[1], commandStrings[2])
	_, err = s.ChannelMessageSend(m.ChannelID, "Success! React channel join set.")
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func removeReactJoinCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		messageExists bool
		validEmoji    = false

		messageID     string
		emojiRegexAPI []string
		emojiAPI      []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 3)

	if len(commandStrings) != 3 && len(commandStrings) != 2 {
		// Returns if the bot called the func
		if m.Author.ID == s.State.User.ID {
			return
		}

		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"removereact [messageID] Optional[emoji]`")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
		// Returns if the bot called the func
		if m.Author.ID == s.State.User.ID {
			return
		}

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid messageID.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	misc.MapMutex.Lock()
	if len(misc.GuildMap[m.GuildID].ReactJoinMap) == 0 {
		// Returns if the bot called the func
		if m.Author.ID == s.State.User.ID {
			misc.MapMutex.Unlock()
			return
		}

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set react joins.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
	for k := range misc.GuildMap[m.GuildID].ReactJoinMap {
		if commandStrings[1] == k {
			messageExists = true
			messageID = k
			break
		}
	}
	misc.MapMutex.Unlock()
	if messageExists == false {
		// Returns if the bot called the func
		if m.Author.ID == s.State.User.ID {
			return
		}

		_, err = s.ChannelMessageSend(m.ChannelID, "Error: No such messageID is set in storage")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
		delete(misc.GuildMap[m.GuildID].ReactJoinMap, commandStrings[1])
		_ = misc.ReactJoinWrite(misc.GuildMap[m.GuildID].ReactJoinMap, m.GuildID)

		// Returns if the bot called the func
		if m.Author.ID == s.State.User.ID {
			misc.MapMutex.Unlock()
			return
		}
		_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed entire message emoji react join.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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

	if misc.GuildMap[m.GuildID].ReactJoinMap[messageID].RoleEmojiMap == nil {
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
	for storageMessageID := range misc.GuildMap[m.GuildID].ReactJoinMap[messageID].RoleEmojiMap {
		for role, emojiSlice := range misc.GuildMap[m.GuildID].ReactJoinMap[messageID].RoleEmojiMap[storageMessageID] {
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
					if len(misc.GuildMap[m.GuildID].ReactJoinMap[messageID].RoleEmojiMap[storageMessageID]) == 1 && len(misc.GuildMap[m.GuildID].ReactJoinMap[messageID].RoleEmojiMap[storageMessageID][role]) == 1 {
						delete(misc.GuildMap[m.GuildID].ReactJoinMap, commandStrings[1])
						_ = misc.ReactJoinWrite(misc.GuildMap[m.GuildID].ReactJoinMap, m.GuildID)

						// Returns if the bot called the func
						if m.Author.ID == s.State.User.ID {
							misc.MapMutex.Unlock()
							return
						}
						_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed emoji react join from message.")
						if err != nil {
							_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
							if err != nil {
								misc.MapMutex.Unlock()
								return
							}
							misc.MapMutex.Unlock()
							return
						}
						// Delete only the role from map if other set react join roles exist in the map
					} else if len(misc.GuildMap[m.GuildID].ReactJoinMap[messageID].RoleEmojiMap[storageMessageID][role]) == 1 {
						delete(misc.GuildMap[m.GuildID].ReactJoinMap[messageID].RoleEmojiMap[storageMessageID], role)
						_ = misc.ReactJoinWrite(misc.GuildMap[m.GuildID].ReactJoinMap, m.GuildID)

						// Returns if the bot called the func
						if m.Author.ID == s.State.User.ID {
							misc.MapMutex.Unlock()
							return
						}
						_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed emoji react join from message.")
						if err != nil {
							_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
							if err != nil {
								misc.MapMutex.Unlock()
								return
							}
							misc.MapMutex.Unlock()
							return
						}
						// Delete only that specific emoji for that specific role
					} else {
						a := misc.GuildMap[m.GuildID].ReactJoinMap[commandStrings[1]].RoleEmojiMap[storageMessageID][role]
						a = append(a[:index], a[index+1:]...)
						misc.GuildMap[m.GuildID].ReactJoinMap[commandStrings[1]].RoleEmojiMap[storageMessageID][role] = a

						// Returns if the bot called the func
						if m.Author.ID == s.State.User.ID {
							misc.MapMutex.Unlock()
							return
						}
						_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed emoji react join from message.")
						if err != nil {
							_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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

		// Returns if the bot called the func
		if m.Author.ID == s.State.User.ID {
			return
		}
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Invalid emoji. Please input a valid emoji or emoji API name.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	if len(misc.GuildMap[m.GuildID].ReactJoinMap) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set react joins.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
	for messageID, value := range misc.GuildMap[m.GuildID].ReactJoinMap {

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
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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

// Saves the react channel join and parses if it already exists
func SaveReactJoin(messageID string, role string, emoji string, guildID string) {

	var (
		temp        misc.ReactJoin
		emojiExists = false
	)

	// Uses this if the message already has a set emoji react
	misc.MapMutex.Lock()
	if misc.GuildMap[guildID].ReactJoinMap[messageID] != nil {
		temp = *misc.GuildMap[guildID].ReactJoinMap[messageID]

		if temp.RoleEmojiMap == nil {
			temp.RoleEmojiMap = append(temp.RoleEmojiMap, misc.GuildMap[guildID].EmojiRoleMap)
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

		misc.GuildMap[guildID].ReactJoinMap[messageID] = &temp
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

	misc.GuildMap[guildID].ReactJoinMap[messageID] = &temp
	misc.MapMutex.Unlock()
}

// Adds role to the user that uses this command if the role is between opt-in dummy roles
func joinCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		roleID      string
		name        string
		chanMention string
		topic       string

		hasRoleAlready bool
		roleExists     bool
	)

	// Pulls info on message author
	mem, err := s.State.Member(m.GuildID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}
	}

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"join [channel/role]`")
		if err != nil {
			_, err := s.ChannelMessageSend(guildPrefix, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls the role name from strings after "joinchannel " or "join "
	if strings.HasPrefix(messageLowercase, guildPrefix+"joinchannel ") {
		name = strings.Replace(messageLowercase, guildPrefix+"joinchannel ", "", -1)
	} else {
		name = strings.Replace(messageLowercase, guildPrefix+"join ", "", -1)
	}

	// Pulls info on server roles
	deb, err := s.GuildRoles(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Pulls info on server channels
	cha, err := s.GuildChannels(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Checks if there's a # before the channel name and removes it if so
	if strings.Contains(name, "#") {
		name = strings.Replace(name, "#", "", -1)

		// Checks if it's in a mention format. If so then user already has access to channel
		if strings.Contains(name, "<") {

			// Fetches mention
			name = strings.Replace(name, ">", "", -1)
			name = strings.Replace(name, "<", "", -1)
			name = misc.ChMentionID(name)

			// Sends error message to user in DMs
			dm, err := s.UserChannelCreate(m.Author.ID)
			if err != nil {
				return
			}
			_, _ = s.ChannelMessageSend(dm.ID, "You're already in "+name)
			return
		}
	}

	// Checks if the role exists on the server, sends error message if not
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == name {
			roleID = deb[i].ID
			if strings.Contains(deb[i].ID, roleID) {
				roleExists = true
				break
			}
		}
	}
	if !roleExists {

		// Sends error message to user in DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, "There's no "+name)
		return
	}

	// Sets role ID
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == name && roleID != "" {
			roleID = deb[i].ID
			break
		}
	}

	// Checks if the user already has the role. Sends error message if he does
	for i := 0; i < len(mem.Roles); i++ {
		if strings.Contains(mem.Roles[i], roleID) {
			hasRoleAlready = true
			break
		}
	}
	if hasRoleAlready {
		// Sets the channel mention to the variable chanMention
		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				chanMention = misc.ChMention(cha[j])
				break
			}
		}

		// Sends error message to user in DMs
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, "You're already in "+chanMention)
		return
	}

	// Confirms whether optins exist
	err = misc.OptInsHandler(s, m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Updates the position of opt-in-under and opt-in-above position
	misc.MapMutex.Lock()
	for i := 0; i < len(deb); i++ {
		if deb[i].ID == misc.GuildMap[m.GuildID].GuildConfig.OptInUnder.ID {
			misc.GuildMap[m.GuildID].GuildConfig.OptInUnder.Position = deb[i].Position
		} else if deb[i].ID == misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.ID {
			misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.Position = deb[i].Position
		}
	}
	misc.MapMutex.Unlock()

	// Sets role
	role, err := s.State.Role(m.GuildID, roleID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Gives role to user if the role is between dummy opt-ins
	misc.MapMutex.Lock()
	if role.Position < misc.GuildMap[m.GuildID].GuildConfig.OptInUnder.Position &&
		role.Position > misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.Position {
		misc.MapMutex.Unlock()
		err = s.GuildMemberRoleAdd(m.GuildID, m.Author.ID, roleID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}

		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				topic = cha[j].Topic
				// Sets the channel mention to the variable chanMention
				chanMention = misc.ChMention(cha[j])
				break
			}
		}

		// Sets DM message
		success := "You have joined "
		if chanMention == "" {
			success += role.Name
		} else {
			success += chanMention
		}
		if topic != "" {
			success = success + "\n **Topic:** " + topic
		}

		// Sends success message to user in DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, success)
		return
	}
	misc.MapMutex.Unlock()
}

// Removes a role from the user that uses this command if the role is between opt-in dummy roles
func leaveCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		roleID      string
		name        string
		chanMention string

		hasRoleAlready bool
		roleExists     bool
	)

	// Pulls info on message author
	mem, err := s.State.Member(m.GuildID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}
	}

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"leave [channel/role]`")
		if err != nil {
			_, err = s.ChannelMessageSend(guildPrefix, err.Error())
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Pulls the role name from strings after "leavechannel " or "leave "
	if strings.HasPrefix(messageLowercase, guildPrefix+"leavechannel ") {
		name = strings.Replace(messageLowercase, guildPrefix+"leavechannel ", "", -1)
	} else {
		name = strings.Replace(messageLowercase, guildPrefix+"leave ", "", -1)
	}

	// Pulls info on server roles
	deb, err := s.GuildRoles(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Pulls info on server channels
	cha, err := s.GuildChannels(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Checks if there's a # before the channel name and removes it if so
	if strings.Contains(name, "#") {
		name = strings.Replace(name, "#", "", -1)
		// Checks if it's in a mention format. If so then user already has access to channel
		if strings.Contains(name, "<") {

			// Fetches mention
			name = strings.Replace(name, ">", "", -1)
			name = strings.Replace(name, "<", "", -1)
			name = misc.ChMentionID(name)

			// Sends error message to user in DMs
			dm, err := s.UserChannelCreate(m.Author.ID)
			if err != nil {
				return
			}
			_, _ = s.ChannelMessageSend(dm.ID, "You cannot leave "+name+" using this command.")
			return
		}
	}

	// Checks if the role exists on the server, sends error message if not
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == name {
			roleID = deb[i].ID
			if strings.Contains(deb[i].ID, roleID) {
				roleExists = true
				break
			}
		}
	}
	if !roleExists {
		// Sends error message to user in DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, "There's no "+name+"")
		return
	}

	// Sets role ID
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == name && roleID != "" {
			roleID = deb[i].ID
			break
		}
	}

	// Checks if the user already has the role. Sends error message if he does
	for i := 0; i < len(mem.Roles); i++ {
		if strings.Contains(mem.Roles[i], roleID) {
			hasRoleAlready = true
			break
		}
	}
	if !hasRoleAlready {

		// Sets the channel mention to the variable chanMention
		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				chanMention = misc.ChMention(cha[j])
				break
			}
		}

		// Sends error message to user in DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, "You're already out of "+chanMention+"")
		return
	}

	// Confirms whether optins exist
	err = misc.OptInsHandler(s, m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Updates the position of opt-in-under and opt-in-above position
	misc.MapMutex.Lock()
	for i := 0; i < len(deb); i++ {
		if deb[i].ID == misc.GuildMap[m.GuildID].GuildConfig.OptInUnder.ID {
			misc.GuildMap[m.GuildID].GuildConfig.OptInUnder.Position = deb[i].Position
		} else if deb[i].ID == misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.ID {
			misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.Position = deb[i].Position
		}
	}
	misc.MapMutex.Unlock()

	// Sets role
	role, err := s.State.Role(m.GuildID, roleID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Removes role from user if the role is between dummy opt-ins
	misc.MapMutex.Lock()
	if role.Position < misc.GuildMap[m.GuildID].GuildConfig.OptInUnder.Position &&
		role.Position > misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.Position {
		misc.MapMutex.Unlock()

		var (
			chanMention string
		)

		err = s.GuildMemberRoleRemove(m.GuildID, m.Author.ID, roleID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}

		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				// Sets the channel mention to the variable chanMention
				chanMention = misc.ChMention(cha[j])
				break
			}
		}

		// Sets DM message
		success := "You have left "
		if chanMention == "" {
			success += role.Name
		} else {
			success += chanMention
		}

		// Sends success message to user in DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, success)
		return
	}
	misc.MapMutex.Unlock()
}

func init() {
	add(&command{
		execute:  setReactJoinCommand,
		trigger:  "setreact",
		aliases:  []string{"setreactjoin", "addreact"},
		desc:     "Sets a react join on a specific message, role and emote. [REACTS]",
		elevated: true,
		category: "reacts",
	})
	add(&command{
		execute:  removeReactJoinCommand,
		trigger:  "removereact",
		aliases:  []string{"removereactjoin", "deletereact"},
		desc:     "Removes a set react join. [REACTS]",
		elevated: true,
		category: "reacts",
	})
	add(&command{
		execute:  viewReactJoinsCommand,
		trigger:  "viewreacts",
		aliases:  []string{"viewreactjoins", "viewreact", "viewreacts", "reacts", "react"},
		desc:     "Views all set react joins. [REACTS]",
		elevated: true,
		category: "reacts",
	})
	add(&command{
		execute:     joinCommand,
		trigger:     "join",
		aliases:     []string{"joinchannel"},
		desc:        "Join a spoiler channel.",
		deleteAfter: true,
		category:    "normal",
	})
	add(&command{
		execute:     leaveCommand,
		trigger:     "leave",
		aliases:     []string{"leavechannel"},
		desc:        "Leave a spoiler channel.",
		deleteAfter: true,
		category:    "normal",
	})
}
