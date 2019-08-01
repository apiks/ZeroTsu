package commands

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

var (
	spamFilterMap      = make(map[string]int)
	spamFilterIsBroken bool
)

// Handles filter in an onMessage basis
func FilterHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	var (
		mLowercase    string
		badWordsSlice []string
		badWordExists bool
		removals      string
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in FilterHandler")
		}
	}()

	if m.GuildID == "" || m.Author == nil {
		return
	}

	// Checks if it's the bot that sent the message
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Pulls info on message author
	mem, err := s.State.Member(m.GuildID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}
	}
	// Checks if user is mod or bot before checking the message
	if HasElevatedPermissions(s, mem.User.ID, m.GuildID) {
		return
	}

	mLowercase = strings.ToLower(m.Content)

	// Checks if message should be filtered
	badWordExists, badWordsSlice = isFiltered(s, m.Message)

	// Exit func if no filtered phrase found
	if !badWordExists {
		return
	}

	// Deletes the message first
	err = s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		return
	}

	if badWordsSlice == nil {
		return
	}

	// Iterates through all the bad words in order and formats print string
	for _, badWord := range badWordsSlice {
		if len(removals) == 0 {
			removals = badWord
		} else {
			removals += ", " + badWord
		}
	}

	// Sends embed mod message
	err = FilterEmbed(s, m.Message, removals, m.ChannelID)
	if err != nil {

		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()

		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Sends message to user's DMs if possible
	dm, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		return
	}
	_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message `%v` was removed for using: _%v_ \n\n", mLowercase, removals))
}

// Handles filter in an onEdit basis
func FilterEditHandler(s *discordgo.Session, m *discordgo.MessageUpdate) {

	var (
		mLowercase    string
		badWordsSlice []string
		badWordExists bool
		removals      string
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in FilterEditHandler")
		}
	}()

	if m.GuildID == "" || m.Author == nil {
		return
	}

	// Checks if it's the bot that sent the message
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Pulls info on message author
	mem, err := s.State.Member(m.GuildID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}
	}
	// Checks if user is mod or bot before checking the message
	if HasElevatedPermissions(s, mem.User.ID, m.GuildID) {
		return
	}

	mLowercase = strings.ToLower(m.Content)

	// Checks if the message should be filtered
	badWordExists, badWordsSlice = isFiltered(s, m.Message)

	// Exit func if no filtered phrase found
	if !badWordExists {
		return
	}

	// Deletes the message first
	err = s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		return
	}

	if badWordsSlice == nil {
		return
	}

	// Iterates through all the bad words in order and formats print string
	for _, badWord := range badWordsSlice {
		if len(removals) == 0 {
			removals = badWord
		} else {
			removals += ", " + badWord
		}
	}

	// Sends embed mod message
	err = FilterEmbed(s, m.Message, removals, m.ChannelID)
	if err != nil {

		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()

		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	// Sends message to user's DMs if possible
	dm, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		return
	}
	_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message `%v` was removed for using: _%v_ \n\n", mLowercase, removals))
}

// Filters reactions that contain a filtered phrase
func FilterReactsHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	var badReactExists bool

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in FilterReactsHandler")
		}
	}()

	if r.GuildID == "" || r.UserID == "" {
		return
	}

	// Checks if it's the bot that sent the message
	if r.UserID == s.State.User.ID {
		return
	}
	// Pulls info on message author
	mem, err := s.State.Member(r.GuildID, r.UserID)
	if err != nil {
		mem, err = s.GuildMember(r.GuildID, r.UserID)
		if err != nil {
			return
		}
	}
	// Checks if user is mod or bot before checking the message
	if HasElevatedPermissions(s, mem.User.ID, r.GuildID) {
		return
	}

	// Checks if the react should be filtered
	badReactExists = isFilteredReact(s, r)

	// Exit func if no filtered phrase found
	if !badReactExists {
		return
	}

	// Deletes the reaction that was sent if it has a filtered phrase
	err = s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.APIName(), r.UserID)
	if err != nil {

		misc.MapMutex.Lock()
		guildBotLog := misc.GuildMap[r.GuildID].GuildConfig.BotLog.ID
		misc.MapMutex.Unlock()

		_, _ = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
	}
}

// Checks if the message is supposed to be filtered
func isFiltered(s *discordgo.Session, m *discordgo.Message) (bool, []string) {

	var (
		mLowercase string
		mentions   string
		userID     string

		mentionCheck []string

		badPhraseSlice         []string
		badPhraseCheckMentions []string
		badPhraseCheck         []string

		messRequireCheckMentions []string
		messRequireCheck         []string
	)

	mLowercase = strings.ToLower(m.Content)

	// Checks if the message contains a mention and finds the actual name instead of ID and put it in mentions
	if strings.Contains(mLowercase, "<@") {

		// Checks for both <@! and <@ mentions
		mentionRegex := regexp.MustCompile(`(?m)<@!?\d+>`)
		mentionCheck = mentionRegex.FindAllString(mLowercase, -1)
		if mentionCheck != nil {
			for _, mention := range mentionCheck {
				userID = strings.TrimPrefix(mention, "<@")
				userID = strings.TrimPrefix(userID, "!")
				userID = strings.TrimSuffix(userID, ">")

				// Checks first in memberInfo. Only checks serverside if it doesn't exist. Saves performance
				misc.MapMutex.Lock()
				if len(misc.GuildMap[m.GuildID].MemberInfoMap) != 0 {
					if _, ok := misc.GuildMap[m.GuildID].MemberInfoMap[userID]; ok {
						mentions += " " + strings.ToLower(misc.GuildMap[m.GuildID].MemberInfoMap[userID].Nickname)
						misc.MapMutex.Unlock()
						continue
					}
				}
				misc.MapMutex.Unlock()

				// If user wasn't found in memberInfo with that username+discrim combo then fetch manually from Discord
				user, err := s.State.Member(m.GuildID, userID)
				if err != nil {
					user, _ = s.GuildMember(m.GuildID, userID)
				}
				if user != nil {
					mentions += " " + strings.ToLower(user.Nick)
				}
			}
		}
	}

	// Iterates through all the filters to see if the message contained a filtered phrase
	misc.MapMutex.Lock()
	for _, filter := range misc.GuildMap[m.GuildID].Filters {

		// Regex check the filter phrase in the message
		re := regexp.MustCompile(filter.Filter)
		badPhraseCheck = re.FindAllString(mLowercase, -1)
		badPhraseCheckMentions = re.FindAllString(mentions, -1)

		// Add all bad phrases in the message if they exist to the slice
		if badPhraseCheck != nil {
			for _, badPhrase := range badPhraseCheck {
				badPhraseSlice = append(badPhraseSlice, badPhrase)
			}
		}
		// Add all bad phrases in the mentions if they exist to the slice
		if badPhraseCheckMentions != nil {
			for _, badMention := range badPhraseCheckMentions {
				badPhraseSlice = append(badPhraseSlice, badMention)
			}
		}
	}
	misc.MapMutex.Unlock()

	// If a bad phrase exists return true to filter it
	if len(badPhraseSlice) != 0 {
		return true, badPhraseSlice
	}

	// Iterates through all of the message requirements to see if the message follows a set requirement {
	misc.MapMutex.Lock()
	for i, requirement := range misc.GuildMap[m.GuildID].MessageRequirements {
		if requirement.Channel != m.ChannelID {
			continue
		}

		// Regex check the requirement phrase in the message
		re := regexp.MustCompile(requirement.Phrase)
		messRequireCheck = re.FindAllString(mLowercase, -1)
		messRequireCheckMentions = re.FindAllString(mentions, -1)

		// If a required phrase exists in the message or mentions, check if it should be removed
		if messRequireCheck != nil {
			misc.GuildMap[m.GuildID].MessageRequirements[i].LastUserID = m.Author.ID
			continue
		}
		if messRequireCheckMentions != nil {
			misc.GuildMap[m.GuildID].MessageRequirements[i].LastUserID = m.Author.ID
			continue
		}

		if requirement.Type == "soft" {
			if requirement.LastUserID == "" {
				misc.GuildMap[m.GuildID].MessageRequirements[i].LastUserID = m.Author.ID
			} else if requirement.LastUserID != m.Author.ID {
				misc.MapMutex.Unlock()
				return true, nil
			}
		}
		if requirement.Type == "hard" {
			misc.MapMutex.Unlock()
			return true, nil
		}
	}
	misc.MapMutex.Unlock()

	return false, nil
}

// Checks if the React is supposed to be filtered
func isFilteredReact(s *discordgo.Session, r *discordgo.MessageReactionAdd) bool {

	var reactName string

	// Iterates through all the filters to see if the react contained a filtered phrase
	misc.MapMutex.Lock()
	for _, filter := range misc.GuildMap[r.GuildID].Filters {

		// Assigns the filter to a string that can be changed to the normal API mode name later
		reactName = filter.Filter

		// Trims the fluff from the filter/reactName (which is a react usually) so it can measured against the API version
		if strings.Contains(reactName, "<:") {
			reactName = strings.Replace(reactName, "<:", "", -1)
			reactName = strings.TrimSuffix(reactName, ">")

		} else if strings.Contains(reactName, "<a:") {
			reactName = strings.Replace(reactName, "<a:", "", -1)
			reactName = strings.TrimSuffix(reactName, ">")
		}

		// Regex check the phrase in the emoji's API name
		re := regexp.MustCompile(reactName)
		badWordCheck := re.FindAllString(r.Emoji.APIName(), -1)
		if badWordCheck == nil {
			continue
		}

		misc.MapMutex.Unlock()
		return true
	}
	misc.MapMutex.Unlock()

	return false
}

// Adds a filter phrase to storage and memory
func addFilterCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		mLowercase     string
		commandStrings []string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	mLowercase = strings.ToLower(m.Content)
	commandStrings = strings.SplitN(mLowercase, " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vfilter [phrase]`\n\n[phrase] is either regex expression (preferable) or just a simple string.", guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Writes the phrase to filters.json and checks if the requirement was already in storage
	err := misc.FiltersWrite(commandStrings[1], m.GuildID)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%v` has been added to the filter list.", commandStrings[1]))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Removes a filter phrase from storage and memory
func removeFilterCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		mLowercase     string
		commandStrings []string
	)

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	if len(misc.GuildMap[m.GuildID].Filters) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no filters.")
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

	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	misc.MapMutex.Unlock()

	mLowercase = strings.ToLower(m.Content)
	commandStrings = strings.SplitN(mLowercase, " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vunfilter [phrase]`\n\n[phrase] is the filter phrase that was used when creating a filter.", guildPrefix))
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Removes phrase from storage and memory
	err := misc.FiltersRemove(commandStrings[1], m.GuildID)
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%v` has been removed from the filter list.", commandStrings[1]))
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Print filters from memory in chat
func viewFiltersCommand(s *discordgo.Session, m *discordgo.Message) {

	var filters string

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	if len(misc.GuildMap[m.GuildID].Filters) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no filters.")
		if err != nil {
			_, err := s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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

	// Iterates through all the filters in memory and adds them to the filters string
	for _, filter := range misc.GuildMap[m.GuildID].Filters {
		if filters == "" {
			filters = fmt.Sprintf("`%v`", filter.Filter)
			continue
		}
		filters = fmt.Sprintf("%v\n `%v`", filters, filter)
	}
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, filters)
	if err != nil {
		_, err := s.ChannelMessageSend(guildBotLog, err.Error())
		if err != nil {
			return
		}
		return
	}
}

func FilterEmbed(s *discordgo.Session, m *discordgo.Message, removals, channelID string) error {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail

		// Embed slice and its fields
		embedField        []*discordgo.MessageEmbedField
		embedFieldFilter  discordgo.MessageEmbedField
		embedFieldMessage discordgo.MessageEmbedField
		embedFieldChannel discordgo.MessageEmbedField

		content string
	)

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	// Checks if the message contains a mention and replaces it with the actual nick instead of ID
	content = m.Content
	content = misc.MentionParser(s, content, m.GuildID)

	// Sets timestamp for removal
	t := time.Now()
	now := t.Format(time.RFC3339)
	embedMess.Timestamp = now

	// Saves user avatar as thumbnail
	embedThumbnail.URL = m.Author.AvatarURL("128")

	// Sets field titles
	embedFieldFilter.Name = "Filtered:"
	embedFieldMessage.Name = "Message:"
	embedFieldChannel.Name = "Channel:"

	// Sets field content
	embedFieldFilter.Value = fmt.Sprintf("**%v**", removals)
	embedFieldMessage.Value = fmt.Sprintf("`%v`", content)
	embedFieldChannel.Value = misc.ChMentionID(channelID)

	// Sets field inline
	embedFieldFilter.Inline = true
	embedFieldChannel.Inline = true

	// Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldFilter)
	embedField = append(embedField, &embedFieldChannel)
	embedField = append(embedField, &embedFieldMessage)

	// Sets embed title and its description (which it uses the same way as a field)
	embedMess.Title = "User:"
	embedMess.Description = m.Author.Mention()

	// Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	// Sends embed in bot-log channel
	_, err := s.ChannelMessageSendEmbed(guildBotLog, &embedMess)
	if err != nil {
		return err
	}

	return nil
}

// Adds a message requirement phrase to storage and memory
func addMessRequirementCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		mLowercase      string
		commandStrings  []string
		channelID       string
		requirementType string
		phrase          string
	)

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	mLowercase = strings.ToLower(m.Content)
	commandStrings = strings.SplitN(mLowercase, " ", 4)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vmrequire [channel]* [type]* [phrase]`\n\n"+
			"`[channel]` is a ping or ID to the channel where the requirement will only be done.\n"+
			"`[type]` can either be soft or hard. Soft means a user must mention the phrase in their first message and is okay until someone else types a message. Hard means all messages must contain that phrase. Defaults to soft.\n"+
			"`[phrase]` is either regex expression (preferable) or just a simple string.\n\n"+
			"***** is optional.", guildPrefix))
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		return
	}
	// Resolves optional parameters based on commandStrings length
	if len(commandStrings) == 2 {
		phrase = commandStrings[1]
	} else if len(commandStrings) == 3 {
		channelID, _ = misc.ChannelParser(s, commandStrings[1], m.GuildID)
		if channelID == "" {
			if commandStrings[1] == "soft" ||
				commandStrings[1] == "hard" {

				requirementType = commandStrings[1]
				phrase = commandStrings[2]
			} else {
				phrase = commandStrings[1] + " " + commandStrings[2]
			}
		} else {
			phrase = commandStrings[2]
		}
	} else if len(commandStrings) == 4 {
		channelID, _ = misc.ChannelParser(s, commandStrings[1], m.GuildID)
		if channelID == "" {
			if commandStrings[1] == "soft" ||
				commandStrings[1] == "hard" {

				requirementType = commandStrings[1]
				phrase = commandStrings[2] + " " + commandStrings[3]
			} else {
				phrase = commandStrings[1] + " " + commandStrings[2] + " " + commandStrings[3]
			}
		} else if commandStrings[2] == "soft" ||
			commandStrings[2] == "hard" {

			requirementType = commandStrings[2]
			phrase = commandStrings[3]
		} else {
			phrase = commandStrings[2] + " " + commandStrings[3]
		}
	}
	if requirementType == "" {
		requirementType = "soft"
	}

	// Writes the phrase to messrequirement.json and checks if the requirement was already in storage
	err := misc.MessRequirementWrite(phrase, channelID, requirementType, m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%v` has been added to the message requirement list.", phrase))
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Removes a message requirement from storage and memory
func removeMessRequirementCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		mLowercase     string
		commandStrings []string
		channelID      string
		phrase         string
	)

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	if len(misc.GuildMap[m.GuildID].MessageRequirements) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no message requirements.")
		if err != nil {
			_, err := s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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

	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	misc.MapMutex.Unlock()

	mLowercase = strings.ToLower(m.Content)
	commandStrings = strings.SplitN(mLowercase, " ", 3)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vunmrequire [channel]* [phrase]`\n\n[channel] is the channel for which that message requirement was set.\n"+
			"`[phrase]` is the phrase that was used when creating a message requirement.\n\n ***** are optional.", guildPrefix))
		if err != nil {
			_, err := s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Resolves optional parameter
	if len(commandStrings) == 3 {
		channelID, _ = misc.ChannelParser(s, commandStrings[1], m.GuildID)
		if channelID == "" {
			phrase = commandStrings[1] + " " + commandStrings[2]
		} else {
			phrase = commandStrings[2]
		}
	} else {
		phrase = commandStrings[1]
	}

	// Removes the phrase from storage and memory
	err := misc.MessRequirementRemove(phrase, channelID, m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%v` has been removed from the message requirement list.", phrase))
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Print message requirements from memory in chat
func viewMessRequirementCommand(s *discordgo.Session, m *discordgo.Message) {

	var mRequirements string

	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	if len(misc.GuildMap[m.GuildID].MessageRequirements) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no message requirements.")
		if err != nil {
			_, err := s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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

	// Iterates through all the message requirements in memory and adds them to the mRequirements string
	for _, requirement := range misc.GuildMap[m.GuildID].MessageRequirements {
		if requirement.Channel == "" {
			requirement.Channel = "All channels"
		}
		mRequirements += fmt.Sprintf("`%v - %v - %v`\n", requirement.Phrase, requirement.Channel, requirement.Type)
	}
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, mRequirements)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
}

// Removes user message if sent too quickly in succession
func SpamFilter(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Checks if the bot had thrown an error before and stops it if so. Helps with massive backlog or delays but disables spam filter
	if spamFilterIsBroken {
		return
	}
	// Stops double event bug with empty content
	if m.Content == "" {
		return
	}
	// Checks if it's the bot that sent the message
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Pulls info on message author
	mem, err := s.State.Member(m.GuildID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			return
		}
	}
	// Checks if user is mod or bot before checking the message
	if HasElevatedPermissions(s, mem.User.ID, m.GuildID) {
		return
	}

	// Counter for how many rapidly sent user messages a user has
	misc.MapMutex.Lock()
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID

	if spamFilterMap[m.Author.ID] < 4 {
		spamFilterMap[m.Author.ID]++
		misc.MapMutex.Unlock()
		return
	}

	// Stops filter if there is an unusual high size of messages to be deleted from a user (i.e. discord lags)
	if spamFilterMap[m.Author.ID] > 15 {
		_, err = s.ChannelMessageSend(guildBotLog, "Error: Spam filter has been disabled due to massive overflow of requests.")
		if err != nil {
			spamFilterIsBroken = true
			misc.MapMutex.Unlock()
			return
		}
		spamFilterIsBroken = true
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Deletes the message if over 4 rapidly sent messages
	err = s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		return
	}
}

// Handles expiring user spam filter map counter
func SpamFilterTimer(s *discordgo.Session, e *discordgo.Ready) {
	for range time.NewTicker(4 * time.Second).C {
		misc.MapMutex.Lock()
		for userID := range spamFilterMap {
			if spamFilterMap[userID] > 0 {
				spamFilterMap[userID]--
			}
		}
		misc.MapMutex.Unlock()
	}
}

// Filters images from the images folder with a tolerance level of 25k
func ImageFilter(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Fetches any image links containing .png
	//urlRegex := regexp.MustCompile(`(?mi)(http[a-zA-Z]?://+)?.+/.+.png`)
	//urls := urlRegex.FindAllString(m.Content, -1)
}

// Adds filter commands to the commandHandler
func init() {
	add(&command{
		execute:  viewFiltersCommand,
		trigger:  "filters",
		aliases:  []string{"viewfilters", "viewfilter"},
		desc:     "Prints all current filters.",
		elevated: true,
		category: "filters",
	})
	add(&command{
		execute:  addFilterCommand,
		trigger:  "addfilter",
		aliases:  []string{"filter", "setfilter"},
		desc:     "Adds a phrase to the filters list.",
		elevated: true,
		category: "filters",
	})
	add(&command{
		execute:  removeFilterCommand,
		trigger:  "removefilter",
		aliases:  []string{"deletefilter", "unfilter"},
		desc:     "Removes a phrase from the filters list.",
		elevated: true,
		category: "filters",
	})
	add(&command{
		execute:  viewMessRequirementCommand,
		trigger:  "mrequirements",
		aliases:  []string{"viewmrequirements", "showmrequirements", "messagerequirements", "messagereqirement", "viewmessrequirements", "messrequirements", "mrequirement", "messrequirement"},
		desc:     "Prints all current message requirement filters.",
		elevated: true,
		category: "filters",
	})
	add(&command{
		execute:  addMessRequirementCommand,
		trigger:  "mrequire",
		aliases:  []string{"messrequire", "setmrequire", "setmessrequire", "setmessagerequire", "addmrequire", "messagerequire", "messrequire", "addmessrequire", "addmessagereqyure"},
		desc:     "Adds a phrase to the message requirement list where it will remove messages that do not contain it.",
		elevated: true,
		category: "filters",
	})
	add(&command{
		execute:  removeMessRequirementCommand,
		trigger:  "unmrequire",
		aliases:  []string{"munrequire", "removemrequire", "removemrequirement", "deletemrequire", "deletemrequirement", "unmessrequire", "deletemessrequire", "removemessrequire"},
		desc:     "Removes a phrase from the message requirement list.",
		elevated: true,
		category: "filters",
	})
}
