package commands

import (
	"log"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// ReactJoinHandler gives a specific role to a user if they react
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

// ReactRemoveHandler removes a role from user if they unreact
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

// addReactJoinCommand adds a react autorole per specific message and emote
func addReactJoinCommand(s *discordgo.Session, messageID, emoji string, role *discordgo.Role, channelID, guildID string) string {
	// Checks if it's a valid messageID
	_, err := strconv.Atoi(messageID)
	if err != nil || len(messageID) < 17 {
		return "Error: Invalid message ID."
	}

	// Parses if it's custom emoji or unicode emoji or animated emoji
	re := regexp.MustCompile("<a?:+([a-zA-Z]|[0-9])+:+[0-9]+>")
	emojiRegex := re.FindAllString(strings.ToLower(emoji), 1)
	if emojiRegex != nil {

		// Fetches emoji API name
		re = regexp.MustCompile("([a-zA-Z]|[0-9])+:[0-9]+")
		emojiName := re.FindAllString(emojiRegex[0], 1)[0]

		// Write
		SaveReactJoin(messageID, role.Name, emojiName, guildID)

		// Reacts with the set emote if possible and gives success
		_ = s.MessageReactionAdd(channelID, messageID, emojiName)

		return "Success! Reaction autorole set."
	}

	// Write
	SaveReactJoin(messageID, role.Name, emoji, guildID)

	// Reacts with the set emote if possible
	_ = s.MessageReactionAdd(channelID, messageID, emoji)
	return "Success! Reaction autorole set."
}

// addReactJoinCommandHandler adds a react autorole per specific message and emote
func addReactJoinCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		roleExists bool
		roleName   string
	)

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
	_, err := strconv.Atoi(commandStrings[1])
	if err != nil || len(commandStrings[1]) < 17 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid message ID.")
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
		if strings.ToLower(role.Name) == commandStrings[3] || role.ID == commandStrings[3] {
			roleExists = true
			roleName = role.Name
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
		SaveReactJoin(commandStrings[1], roleName, emojiName, m.GuildID)

		// Reacts with the set emote if possible and gives success
		_ = s.MessageReactionAdd(m.ChannelID, commandStrings[1], emojiName)
		_, err = s.ChannelMessageSend(m.ChannelID, "Success! React channel join set.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

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

func removeReactJoinCommand(messageID, emoji, guildID string) string {
	var (
		messageExists bool
		validEmoji    bool

		emojiRegexAPI []string
		emojiAPI      []string

		guildReactJoin = db.GetGuildReactJoin(guildID)
	)

	// Checks if it's a valid messageID
	_, err := strconv.Atoi(messageID)
	if err != nil || len(messageID) < 17 {
		return "Error: Invalid message ID."
	}

	if len(guildReactJoin) == 0 {
		return "Error: There are no set reaction autoroles."
	}

	// Checks if the messageID already exists in the map
	for i := range guildReactJoin {
		if messageID == i {
			messageExists = true
			messageID = i
			break
		}
	}

	if messageExists == false {
		return "Error: No such message ID is set."
	}

	// Removes the entire message from the map and writes to storage
	if emoji == "" {
		delete(guildReactJoin, messageID)
		db.SetGuildReactJoin(guildID, guildReactJoin)
		return "Success! Removed entire message react autorole."
	}

	if guildReactJoin[messageID].GetRoleEmojiMap() == nil {
		return ""
	}

	// Parses if it's custom emoji or unicode
	re := regexp.MustCompile("(?i)<:+([a-zA-Z]|[0-9])+:+[0-9]+>")
	emojiRegex := re.FindAllString(emoji, 1)
	if emojiRegex == nil {
		// Second parser if it's custom emoji or unicode but for emoji API name instead
		reAPI := regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
		emojiRegexAPI = reAPI.FindAllString(emoji, 1)
	}

	for storageMessageID := range guildReactJoin[messageID].GetRoleEmojiMap() {
		for role, emojiSlice := range guildReactJoin[messageID].GetRoleEmojiMap()[storageMessageID] {
			if emojiSlice == nil {
				continue
			}

			for index, emojiLoop := range emojiSlice {
				// Checks for unicode emoji
				if len(emojiRegex) == 0 && len(emojiRegexAPI) == 0 {
					if emoji == emojiLoop {
						validEmoji = true
					}
					// Checks for non-unicode emoji
				} else {
					// Trims non-unicode emoji name to fit API emoji name
					re = regexp.MustCompile("(?i)([a-zA-Z]|[0-9])+:[0-9]+")
					if len(emojiRegex) == 0 {
						if len(emojiRegexAPI) != 0 {
							emojiAPI = re.FindAllString(emojiRegexAPI[0], 1)
							if emojiLoop == emojiAPI[0] {
								validEmoji = true
							}
						}
					} else {
						emojiAPI = re.FindAllString(emojiRegex[0], 1)
						if emojiLoop == emojiAPI[0] {
							validEmoji = true
						}
					}
				}

				// Delete only if it's a valid emoji in map
				if validEmoji {
					// Delete the entire message from map if it's the only set emoji react join
					if len(guildReactJoin[messageID].GetRoleEmojiMap()[storageMessageID][role]) == 1 {
						delete(guildReactJoin, messageID)
						db.SetGuildReactJoin(guildID, guildReactJoin)
						return "Success! Removed that emoji autorole from the message."

						// Delete only that specific emoji for that specific role
					} else {
						a := guildReactJoin[messageID].GetRoleEmojiMap()[storageMessageID][role]
						a = append(a[:index], a[index+1:]...)
						guildReactJoin[messageID].GetRoleEmojiMap()[storageMessageID][role] = a
						return "Success! Removed that emoji autorole from the message."
					}
				}
			}
		}
	}

	// If it comes this far it means it's an invalid emoji
	if emojiRegex == nil && emojiRegexAPI == nil {
		return "Error: Invalid emoji. Please input a valid emoji or emoji API name."
	}

	return ""
}

func removeReactJoinCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

func viewReactJoinsCommand(guildID string) []string {
	var (
		message           string
		guildReactJoinMap = db.GetGuildReactJoin(guildID)
	)

	if len(guildReactJoinMap) == 0 {
		return []string{"Error: There are no set react joins."}
	}

	// Iterates through all of the set channel joins and assigns them to a string
	for messageID, value := range guildReactJoinMap {
		if value == nil {
			continue
		}

		// Formats message
		message = "——————\n`MessageID: " + (messageID + "`\n")
		for i := 0; i < len(value.GetRoleEmojiMap()); i++ {
			for role, emoji := range value.GetRoleEmojiMap()[i] {
				message = message + "`" + role + "` — "
				for j := 0; j < len(emoji); j++ {
					if j != len(emoji)-1 {
						message = message + emoji[j] + ", "
					} else {
						message = message + emoji[j] + "\n"
					}
				}
			}
		}
	}

	return common.SplitLongMessage(message)
}

func viewReactJoinsCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

// SaveReactJoin saves the react channel join and parses if it already exists
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

func init() {
	Add(&Command{
		Execute:    addReactJoinCommandHandler,
		Name:       "add-react-autorole",
		Aliases:    []string{"setreactjoin", "addreact", "setreact", "set-react", "set-react-autorole"},
		Desc:       "Adds a react autorole on a specific message, emoji and role",
		Permission: functionality.Mod,
		Module:     "reacts",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message-id",
				Description: "The ID of the message you want to set a reaction emoji and role for.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "emoji",
				Description: "The emoji you want to set as a reaction emoji.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "The role you want it to give and take whenever a user reacts with the specified emoji.",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "add-react-autorole", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			var (
				messageID string
				emoji     string
				role      *discordgo.Role
			)
			if i.ApplicationCommandData().Options == nil {
				return
			}

			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "message-id" {
					messageID = option.StringValue()
				} else if option.Name == "emoji" {
					emoji = option.StringValue()
				} else if option.Name == "role" {
					role = option.RoleValue(s, i.GuildID)
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: addReactJoinCommand(s, messageID, emoji, role, i.ChannelID, i.GuildID),
				},
			})
		},
	})
	Add(&Command{
		Execute:    removeReactJoinCommandHandler,
		Name:       "remove-react-autorole",
		Aliases:    []string{"removereactjoin", "deletereact", "removereact", "removereactautorole", "deletereactautorole", "delete-react-autorole"},
		Desc:       "Removes a set react join",
		Permission: functionality.Mod,
		Module:     "reacts",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message-id",
				Description: "The ID of the message you want to remove a react autorole for.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "emoji",
				Description: "The emoji you want to remove as a reaction emoji to that message.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "remove-react-autorole", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			var (
				messageID string
				emoji     string
			)
			if i.ApplicationCommandData().Options == nil {
				return
			}

			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "message-id" {
					messageID = option.StringValue()
				} else if option.Name == "emoji" {
					emoji = option.StringValue()
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: removeReactJoinCommand(messageID, emoji, i.GuildID),
				},
			})
		},
	})
	Add(&Command{
		Execute:    viewReactJoinsCommandHandler,
		Name:       "reacts-autorole",
		Aliases:    []string{"viewreactjoins", "viewreact", "viewreacts", "reacts", "react", "viewreacts", "reactsautorole", "react-autorole"},
		Desc:       "Prints all set reaction autoroles",
		Permission: functionality.Mod,
		Module:     "reacts",
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "reacts-autorole", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			messages := viewReactJoinsCommand(i.GuildID)
			if messages == nil {
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: messages[0],
				},
			})

			if len(messages) > 1 {
				for j, message := range messages {
					if j == 0 {
						continue
					}

					s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
						Content: message,
					})
				}
			}
		},
	})
}
