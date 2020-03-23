package commands

import (
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"log"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Gives a specific role to a user if they react
func ReactJoinHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in ReactJoinHandler")
			log.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()

	if r.GuildID == "" {
		return
	}

	entities.HandleNewGuild(r.GuildID)

	guildReactJoin := db.GetGuildReactJoin(r.GuildID)
	// Checks if a react channel join is set for that specific message and emoji and continues if true
	if guildReactJoin == nil ||
		guildReactJoin[r.MessageID] == nil ||
		guildReactJoin[r.MessageID].GetRoleEmojiMap() == nil {
		return
	}
	guildSettings := db.GetGuildSettings(r.GuildID)
	guildRoleEmojiMap := guildReactJoin[r.MessageID].GetRoleEmojiMap()

	// Return if the one reacting is this BOT
	if r.UserID == s.State.SessionID {
		return
	}

	// Pulls all of the server roles
	roles, err := s.GuildRoles(r.GuildID)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Puts the react API emoji name to lowercase so it is valid with the storage emoji name
	reactLowercase := strings.ToLower(r.Emoji.APIName())

	for _, roleEmojiMap := range guildRoleEmojiMap {
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
							common.LogError(s, guildSettings.BotLog, err)
							return
						}
						return
					}
				}
				for _, serverRole := range roles {
					if strings.ToLower(serverRole.Name) == strings.ToLower(role) {
						// Gives the role
						err := s.GuildMemberRoleAdd(r.GuildID, r.UserID, serverRole.ID)
						if err != nil {
							common.LogError(s, guildSettings.BotLog, err)
							return
						}
					}
				}
			}
		}
	}
}

// Removes a role from user if they unreact
func ReactRemoveHandler(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in ReactRemoveHandler")
			log.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()

	if r.GuildID == "" {
		return
	}

	entities.HandleNewGuild(r.GuildID)

	guildReactJoin := db.GetGuildReactJoin(r.GuildID)
	// Checks if a react channel join is set for that specific message and emoji and continues if true
	if guildReactJoin == nil ||
		guildReactJoin[r.MessageID] == nil ||
		guildReactJoin[r.MessageID].GetRoleEmojiMap() == nil {
		return
	}
	guildSettings := db.GetGuildSettings(r.GuildID)
	guildRoleEmojiMap := guildReactJoin[r.MessageID].GetRoleEmojiMap()

	// Return if the one unreacting is this BOT
	if r.UserID == s.State.SessionID {
		return
	}

	// Pulls all of the server roles
	roles, err := s.GuildRoles(r.GuildID)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Puts the react API emoji name to lowercase so it is valid with the storage emoji name
	reactLowercase := strings.ToLower(r.Emoji.APIName())

	for _, roleEmojiMap := range guildRoleEmojiMap {
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
						_ = s.GuildMemberRoleRemove(r.GuildID, r.UserID, role)
						return
					}
				}
				for _, serverRole := range roles {
					if strings.ToLower(serverRole.Name) == strings.ToLower(role) {
						// Removes the role
						err := s.GuildMemberRoleRemove(r.GuildID, r.UserID, serverRole.ID)
						if err != nil {
							common.LogError(s, guildSettings.BotLog, err)
							return
						}
					}
				}
			}
		}
	}
}

// Sets react joins per specific message and emote
func setReactJoinCommand(s *discordgo.Session, m *discordgo.Message) {
	var roleExists bool

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 4)

	if len(commandStrings) != 4 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"setreact [messageID] [emoji] [role]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if it's a valid messageID
	num, err := strconv.Atoi(commandStrings[1])
	if err != nil || num < 17 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid messageID.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fetches all server roles
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parses if it's custom emoji or unicode emoji or animated emoji
	re := regexp.MustCompile("<a?:+([a-zA-Z]|[0-9])+:+[0-9]+>")
	emojiRegex := re.FindAllString(strings.ToLower(m.Content), 1)
	if emojiRegex != nil {

		// Fetches emoji API name
		re = regexp.MustCompile("([a-zA-Z]|[0-9])+:[0-9]+")
		emojiName := re.FindAllString(emojiRegex[0], 1)[0]

		// Write
		SaveReactJoin(commandStrings[1], commandStrings[3], emojiName, m.GuildID)

		// Reacts with the set emote if possible and gives success
		_ = s.MessageReactionAdd(m.ChannelID, commandStrings[1], emojiName)
		_, err = s.ChannelMessageSend(m.ChannelID, "Success! React channel join set.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// If the above is false, it's a non-valid emoji or an unicode emoji (the latter preferably) and saves that

	// Write
	SaveReactJoin(commandStrings[1], commandStrings[3], commandStrings[2], m.GuildID)

	// Reacts with the set emote if possible
	_ = s.MessageReactionAdd(m.ChannelID, commandStrings[1], commandStrings[2])
	_, err = s.ChannelMessageSend(m.ChannelID, "Success! React channel join set.")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
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

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildReactJoin := db.GetGuildReactJoin(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 3)

	if len(commandStrings) != 3 && len(commandStrings) != 2 {
		// Returns if the bot called the func
		if m.Author.ID == s.State.User.ID {
			return
		}

		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"removereact [messageID] Optional[emoji]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
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
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(guildReactJoin) == 0 {
		if m.Author.ID == s.State.User.ID {
			return
		}
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set react joins.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	// Checks if the messageID already exists in the map
	for k := range guildReactJoin {
		if commandStrings[1] == k {
			messageExists = true
			messageID = k
			break
		}
	}

	if messageExists == false {
		// Returns if the bot called the func
		if m.Author.ID == s.State.User.ID {
			return
		}

		_, err = s.ChannelMessageSend(m.ChannelID, "Error: No such messageID is set in storage")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Removes the entire message from the map and writes to storage
	if len(commandStrings) == 2 {
		delete(guildReactJoin, commandStrings[1])
		db.SetGuildReactJoin(m.GuildID, guildReactJoin)

		// Returns if the bot called the func
		if m.Author.ID == s.State.User.ID {
			return
		}
		_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed entire message emoji react join.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if guildReactJoin[messageID].GetRoleEmojiMap() == nil {
		return
	}

	// Parses if it's custom emoji or unicode
	re := regexp.MustCompile("(?i)<:+([a-zA-Z]|[0-9])+:+[0-9]+>")
	emojiRegex := re.FindAllString(commandStrings[2], 1)
	if emojiRegex == nil {
		// Second parser if it's custom emoji or unicode but for emoji API name instead
		reAPI := regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
		emojiRegexAPI = reAPI.FindAllString(commandStrings[2], 1)
	}

	for storageMessageID := range guildReactJoin[messageID].GetRoleEmojiMap() {
		for role, emojiSlice := range guildReactJoin[messageID].GetRoleEmojiMap()[storageMessageID] {
			if emojiSlice == nil {
				continue
			}

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
					if len(guildReactJoin) == 1 && len(guildReactJoin[messageID].GetRoleEmojiMap()[storageMessageID][role]) == 1 {
						delete(guildReactJoin, commandStrings[1])
						db.SetGuildReactJoin(m.GuildID, guildReactJoin)

						// Returns if the bot called the func
						if m.Author.ID == s.State.User.ID {
							return
						}
						_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed emoji react join from message.")
						if err != nil {
							common.LogError(s, guildSettings.BotLog, err)
							return
						}
						// Delete only the role from map if other set react join roles exist in the map
					} else if len(guildReactJoin[messageID].GetRoleEmojiMap()[storageMessageID][role]) == 1 {
						delete(guildReactJoin[messageID].GetRoleEmojiMap()[storageMessageID], role)
						db.SetGuildReactJoin(m.GuildID, guildReactJoin)

						// Returns if the bot called the func
						if m.Author.ID == s.State.User.ID {
							return
						}
						_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed emoji react join from message.")
						if err != nil {
							common.LogError(s, guildSettings.BotLog, err)
							return
						}
						// Delete only that specific emoji for that specific role
					} else {
						a := guildReactJoin[commandStrings[1]].GetRoleEmojiMap()[storageMessageID][role]
						a = append(a[:index], a[index+1:]...)
						guildReactJoin[commandStrings[1]].GetRoleEmojiMap()[storageMessageID][role] = a

						// Returns if the bot called the func
						if m.Author.ID == s.State.User.ID {
							return
						}
						_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed emoji react join from message.")
						if err != nil {
							common.LogError(s, guildSettings.BotLog, err)
							return
						}
					}
					return
				}

			}
		}
	}

	// If it comes this far it means it's an invalid emoji
	if emojiRegex == nil && emojiRegexAPI == nil {

		// Returns if the bot called the func
		if m.Author.ID == s.State.User.ID {
			return
		}
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Invalid emoji. Please input a valid emoji or emoji API name.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
}

// Prints all currently set React Joins in memory
func viewReactJoinsCommand(s *discordgo.Session, m *discordgo.Message) {

	var line string

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildReactJoinMap := db.GetGuildReactJoin(m.GuildID)

	if len(guildReactJoinMap) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no set react joins.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Iterates through all of the set channel joins and assigns them to a string
	for messageID, value := range guildReactJoinMap {
		if value == nil {
			continue
		}

		// Formats message
		line = "——————\n`MessageID: " + (messageID + "`\n")
		for i := 0; i < len(value.GetRoleEmojiMap()); i++ {
			for role, emoji := range value.GetRoleEmojiMap()[i] {
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
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
}

// Saves the react channel join and parses if it already exists
func SaveReactJoin(messageID string, role string, emoji string, guildID string) {
	var (
		emojiExists bool
		temp        entities.ReactJoin
	)

	// Uses this if the message already has a set emoji react
	guildReactJoin := db.GetGuildReactJoin(guildID)

	if guildReactJoin[messageID] != nil {
		for i := 0; i < len(guildReactJoin[messageID].GetRoleEmojiMap()); i++ {
			if guildReactJoin[messageID].GetRoleEmojiMap()[i][role] == nil {
				guildReactJoin[messageID].GetRoleEmojiMap()[i][role] = append(guildReactJoin[messageID].GetRoleEmojiMap()[i][role], emoji)
			}

			for j := 0; j < len(guildReactJoin[messageID].GetRoleEmojiMap()[i][role]); j++ {
				if guildReactJoin[messageID].GetRoleEmojiMap()[i][role][j] == emoji {
					emojiExists = true
					break
				}
			}
			if !emojiExists {
				guildReactJoin[messageID].GetRoleEmojiMap()[i][role] = append(guildReactJoin[messageID].GetRoleEmojiMap()[i][role], emoji)
			}
		}

		db.SetGuildReactJoinEmoji(guildID, messageID, guildReactJoin[messageID])
		return
	}

	// Initializes temp.RoleEmoji if the message doesn't have a set emoji react
	EmojiRoleMapDummy := make(map[string][]string)
	if temp.GetRoleEmojiMap() == nil {
		temp.AppendToRoleEmojiMap(EmojiRoleMapDummy)
	}

	for i := 0; i < len(temp.GetRoleEmojiMap()); i++ {
		if temp.GetRoleEmojiMap()[i][role] == nil {
			temp.GetRoleEmojiMap()[i][role] = append(temp.GetRoleEmojiMap()[i][role], emoji)
		}
	}

	db.SetGuildReactJoinEmoji(guildID, messageID, &temp)
}

// Adds role to the user that uses this command if the role is between opt-in dummy roles
func joinCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		role *discordgo.Role

		name        string
		chanMention string
		topic       string

		hasRoleAlready bool
	)

	// Pulls info on message author
	mem, err := s.State.Member(m.GuildID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}
	}

	guildSettings := db.GetGuildSettings(m.GuildID)
	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"join [channel/role]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	name = strings.Join(commandStrings[1:], "-")

	// Pulls info on server roles
	deb, err := s.GuildRoles(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Pulls info on server channels
	cha, err := s.GuildChannels(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
			name = common.ChMentionID(name)

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
			role = deb[i]
			break
		}
	}
	if role == nil {
		name = strings.Replace(name, " ", "-", -1)
		for i := 0; i < len(deb); i++ {
			if deb[i].Name == name {
				role = deb[i]
			}
		}
	}

	// Sends error message to user in DMs if possible
	if role == nil {
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, "There's no "+name)
		return
	}

	// Checks if the user already has the role. Sends error message if he does
	for i := 0; i < len(mem.Roles); i++ {
		if strings.Contains(mem.Roles[i], role.ID) {
			hasRoleAlready = true
			break
		}
	}
	if hasRoleAlready {
		// Sets the channel mention to the variable chanMention
		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				chanMention = common.ChMention(cha[j])
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
	err = common.OptInsHandler(s, m.ChannelID, m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Updates the position of opt-in-under and opt-in-above position
	for i := 0; i < len(deb); i++ {
		if deb[i].ID == guildSettings.GetOptInUnder().GetID() {
			optInUnder := guildSettings.GetOptInUnder()
			optInUnder = optInUnder.SetPosition(deb[i].Position)
			guildSettings = guildSettings.SetOptInUnder(optInUnder)

			db.SetGuildSettings(m.GuildID, guildSettings)
		} else if deb[i].ID == guildSettings.GetOptInAbove().GetID() {
			optInAbove := guildSettings.GetOptInAbove()
			optInAbove = optInAbove.SetPosition(deb[i].Position)
			guildSettings = guildSettings.SetOptInAbove(optInAbove)

			db.SetGuildSettings(m.GuildID, guildSettings)
		}
	}

	// Gives role to user if the role is between dummy opt-ins
	if role.Position < guildSettings.GetOptInUnder().GetPosition() &&
		role.Position > guildSettings.GetOptInAbove().GetPosition() {

		err = s.GuildMemberRoleAdd(m.GuildID, m.Author.ID, role.ID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}

		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				topic = cha[j].Topic
				// Sets the channel mention to the variable chanMention
				chanMention = common.ChMention(cha[j])
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
	} else {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: the target role/channel is not set as opt-in (between the dummy opt-in roles).")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}
}

// Removes a role from the user that uses this command if the role is between opt-in dummy roles
func leaveCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		role        *discordgo.Role
		name        string
		chanMention string

		hasRoleAlready bool
	)

	// Pulls info on message author
	mem, err := s.State.Member(m.GuildID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}
	}

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"leave [channel/role]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	name = strings.Join(commandStrings[1:], "-")

	// Pulls info on server roles
	deb, err := s.GuildRoles(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Pulls info on server channels
	cha, err := s.GuildChannels(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
			name = common.ChMentionID(name)

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
			role = deb[i]
			break
		}
	}
	if role == nil {
		name = strings.Replace(name, " ", "-", -1)
		for i := 0; i < len(deb); i++ {
			if deb[i].Name == name {
				role = deb[i]
				break
			}
		}
	}
	if role == nil {
		// Sends error message to user in DMs if possible
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			return
		}
		_, _ = s.ChannelMessageSend(dm.ID, "There's no "+name+"")
		return
	}

	// Checks if the user already has the role. Sends error message if he does
	for i := 0; i < len(mem.Roles); i++ {
		if strings.Contains(mem.Roles[i], role.ID) {
			hasRoleAlready = true
			break
		}
	}
	if !hasRoleAlready {

		// Sets the channel mention to the variable chanMention
		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				chanMention = common.ChMention(cha[j])
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
	err = common.OptInsHandler(s, m.ChannelID, m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Updates the position of opt-in-under and opt-in-above position
	for i := 0; i < len(deb); i++ {
		if deb[i].ID == guildSettings.GetOptInUnder().GetID() {
			optInUnder := guildSettings.GetOptInUnder()
			optInUnder = optInUnder.SetPosition(deb[i].Position)
			guildSettings = guildSettings.SetOptInUnder(optInUnder)

			db.SetGuildSettings(m.GuildID, guildSettings)
		} else if deb[i].ID == guildSettings.GetOptInAbove().GetID() {
			optInAbove := guildSettings.GetOptInAbove()
			optInAbove = optInAbove.SetPosition(deb[i].Position)
			guildSettings = guildSettings.SetOptInAbove(optInAbove)

			db.SetGuildSettings(m.GuildID, guildSettings)
		}
	}

	// Removes role from user if the role is between dummy opt-ins
	if role.Position < guildSettings.GetOptInUnder().GetPosition() &&
		role.Position > guildSettings.GetOptInAbove().GetPosition() {

		var (
			chanMention string
		)

		err = s.GuildMemberRoleRemove(m.GuildID, m.Author.ID, role.ID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}

		for j := 0; j < len(cha); j++ {
			if cha[j].Name == name {
				// Sets the channel mention to the variable chanMention
				chanMention = common.ChMention(cha[j])
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
	} else {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: the target role/channel is not set as opt-in (between the dummy opt-in roles).")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}
}

func init() {
	Add(&Command{
		Execute:    setReactJoinCommand,
		Trigger:    "setreact",
		Aliases:    []string{"setreactjoin", "addreact"},
		Desc:       "Sets a react join on a specific message, role and emote [REACTS]",
		Permission: functionality.Mod,
		Module:     "reacts",
	})
	Add(&Command{
		Execute:    removeReactJoinCommand,
		Trigger:    "removereact",
		Aliases:    []string{"removereactjoin", "deletereact"},
		Desc:       "Removes a set react join [REACTS]",
		Permission: functionality.Mod,
		Module:     "reacts",
	})
	Add(&Command{
		Execute:    viewReactJoinsCommand,
		Trigger:    "viewreacts",
		Aliases:    []string{"viewreactjoins", "viewreact", "viewreacts", "reacts", "react"},
		Desc:       "Views all set react joins [REACTS]",
		Permission: functionality.Mod,
		Module:     "reacts",
	})
	Add(&Command{
		Execute:     joinCommand,
		Trigger:     "join",
		Aliases:     []string{"joinchannel"},
		Desc:        "Join a spoiler channel",
		DeleteAfter: true,
		Module:      "normal",
	})
	Add(&Command{
		Execute:     leaveCommand,
		Trigger:     "leave",
		Aliases:     []string{"leavechannel"},
		Desc:        "Leave a spoiler channel",
		DeleteAfter: true,
		Module:      "normal",
	})
}
