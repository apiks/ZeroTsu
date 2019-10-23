package commands

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

//var (
//	spamFilterMap      = make(map[string]int)
//	spamFilterIsBroken bool
//)

// Handles filter in an onMessage basis
func FilterHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

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

	functionality.HandleNewGuild(s, m.GuildID)

	// Checks if user is mod or bot before checking the message
	if functionality.HasElevatedPermissions(s, m.Author.ID, m.GuildID) {
		return
	}

	var (
		badWordsSlice []string
		badWordExists bool
		removals      string
	)

	// Checks if message should be filtered
	badWordExists, badWordsSlice = isFiltered(s, m.Message)

	// Exit func if no filtered phrase found
	if !badWordExists {
		return
	}

	// Deletes the message first
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
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
	err = functionality.FilterEmbed(s, m.Message, removals, m.ChannelID)
	if err != nil {

		functionality.Mutex.RLock()
		guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.Mutex.RUnlock()

		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Sends message to user's DMs if possible
	dm, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		return
	}
	_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message `%v` was removed for using: _%v_ \n\n", strings.ToLower(m.Content), removals))
}

// Handles filter in an onEdit basis
func FilterEditHandler(s *discordgo.Session, m *discordgo.MessageUpdate) {

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

	functionality.HandleNewGuild(s, m.GuildID)

	// Checks if user is mod or bot before checking the message
	if functionality.HasElevatedPermissions(s, m.Author.ID, m.GuildID) {
		return
	}

	var (
		badWordsSlice []string
		badWordExists bool
		removals      string
	)

	// Checks if the message should be filtered
	badWordExists, badWordsSlice = isFiltered(s, m.Message)

	// Exit func if no filtered phrase found
	if !badWordExists {
		return
	}

	// Deletes the message first
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
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
	err = functionality.FilterEmbed(s, m.Message, removals, m.ChannelID)
	if err != nil {

		functionality.Mutex.RLock()
		guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.Mutex.RUnlock()

		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}

	// Sends message to user's DMs if possible
	dm, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		return
	}
	_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message `%s` was removed for using: _%v_ \n\n", strings.ToLower(m.Content), removals))
}

// Filters reactions that contain a filtered phrase
func FilterReactsHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

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

	functionality.HandleNewGuild(s, r.GuildID)

	// Checks if user is mod or bot before checking the message
	if functionality.HasElevatedPermissions(s, r.UserID, r.GuildID) {
		return
	}

	var badReactExists bool

	// Checks if the react should be filtered
	badReactExists = isFilteredReact(s, r)

	// Exit func if no filtered phrase found
	if !badReactExists {
		return
	}

	// Deletes the reaction that was sent if it has a filtered phrase
	err := s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.APIName(), r.UserID)
	if err != nil {

		functionality.Mutex.RLock()
		guildSettings := functionality.GuildMap[r.GuildID].GetGuildSettings()
		functionality.Mutex.RUnlock()

		functionality.LogError(s, guildSettings.BotLog, err)
		return
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

	if m.GuildID == "" {
		return false, nil
	}

	mLowercase = strings.ToLower(m.Content)

	// Checks if the message contains a mention and finds the actual name instead of ID and put it in mentions
	if strings.Contains(mLowercase, "<@") {

		// Checks for both <@! and <@ mentions
		mentionRegex := regexp.MustCompile(`(?m)<@!?\d+>`)
		mentionCheck = mentionRegex.FindAllString(mLowercase, -1)
		if mentionCheck != nil {
			var wg sync.WaitGroup
			wg.Add(len(mentionCheck))

			for _, mention := range mentionCheck {
				go func(mention string) {
					defer wg.Done()

					userID = strings.TrimPrefix(mention, "<@")
					userID = strings.TrimPrefix(userID, "!")
					userID = strings.TrimSuffix(userID, ">")

					// Checks first in memberInfo. Only checks serverside if it doesn't exist. Saves performance
					functionality.Mutex.RLock()
					if len(functionality.GuildMap[m.GuildID].MemberInfoMap) != 0 {
						if _, ok := functionality.GuildMap[m.GuildID].MemberInfoMap[userID]; ok {
							mentions += " " + strings.ToLower(functionality.GuildMap[m.GuildID].MemberInfoMap[userID].Nickname)
							functionality.Mutex.RUnlock()
							return
						}
					}
					functionality.Mutex.RUnlock()

					// If user wasn't found in memberInfo with that username+discrim combo then fetch manually from Discord
					user, err := s.State.Member(m.GuildID, userID)
					if err != nil {
						user, err = s.GuildMember(m.GuildID, userID)
						if err != nil {
							return
						}
					}
					mentions += fmt.Sprintf(" %s", strings.ToLower(user.Nick))
				}(mention)

			}

			wg.Wait()
		}
	}

	// Iterates through all the filters to see if the message contained a filtered phrase
	functionality.Mutex.RLock()
	guildFilters := functionality.GuildMap[m.GuildID].Filters
	functionality.Mutex.RUnlock()
	for _, filter := range guildFilters {

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

	// If a bad phrase exists return true to filter it
	if len(badPhraseSlice) != 0 {
		return true, badPhraseSlice
	}

	// Iterates through all of the message requirements to see if the message follows a set requirement
	functionality.Mutex.Lock()
	for i, requirement := range functionality.GuildMap[m.GuildID].MessageRequirements {
		if requirement.Channel != m.ChannelID {
			continue
		}

		// Regex check the requirement phrase in the message
		re := regexp.MustCompile(requirement.Phrase)
		messRequireCheck = re.FindAllString(mLowercase, -1)
		messRequireCheckMentions = re.FindAllString(mentions, -1)

		// If a required phrase exists in the message or mentions, check if it should be removed
		if messRequireCheck != nil {
			functionality.GuildMap[m.GuildID].MessageRequirements[i].LastUserID = m.Author.ID
			continue
		}
		if messRequireCheckMentions != nil {
			functionality.GuildMap[m.GuildID].MessageRequirements[i].LastUserID = m.Author.ID
			continue
		}

		if requirement.Type == "soft" {
			if requirement.LastUserID == "" {
				functionality.GuildMap[m.GuildID].MessageRequirements[i].LastUserID = m.Author.ID
			} else if requirement.LastUserID != m.Author.ID {
				functionality.Mutex.Unlock()
				return true, nil
			}
		}
		if requirement.Type == "hard" {
			functionality.Mutex.Unlock()
			return true, nil
		}
	}
	functionality.Mutex.Unlock()

	return false, nil
}

// Checks if the React is supposed to be filtered
func isFilteredReact(s *discordgo.Session, r *discordgo.MessageReactionAdd) bool {

	var reactName string

	// Iterates through all the filters to see if the react contained a filtered phrase
	functionality.Mutex.RLock()
	guildFilters := functionality.GuildMap[r.GuildID].Filters
	functionality.Mutex.RUnlock()

	for _, filter := range guildFilters {

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

		return true
	}

	return false
}

// Adds a filter phrase to storage and memory
func addFilterCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sfilter [phrase]`\n\n[phrase] is either regex expression (preferable) or just a simple string.", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Removes arrows from emojis, mentions and channels
	commandStrings[1] = strings.TrimPrefix(commandStrings[1], "<")
	commandStrings[1] = strings.TrimSuffix(commandStrings[1], ">")

	// Writes the phrase to filters.json and checks if the requirement was already in storage
	err := functionality.FiltersWrite(commandStrings[1], m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%s` has been added to the filter list.", commandStrings[1]))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a filter phrase from storage and memory
func removeFilterCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	filtersLen := len(functionality.GuildMap[m.GuildID].Filters)
	functionality.Mutex.RUnlock()

	if filtersLen == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no filters.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sunfilter [phrase]`\n\n[phrase] is the filter phrase that was used when creating a filter.", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Removes arrows from emojis, mentions and channels
	commandStrings[1] = strings.TrimPrefix(commandStrings[1], "<")
	commandStrings[1] = strings.TrimSuffix(commandStrings[1], ">")

	// Removes phrase from storage and memory
	err := functionality.FiltersRemove(commandStrings[1], m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%s` has been removed from the filter list.", commandStrings[1]))
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Print filters from memory in chat
func viewFiltersCommand(s *discordgo.Session, m *discordgo.Message) {

	var filters string

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	guildFilters := functionality.GuildMap[m.GuildID].Filters
	functionality.Mutex.RUnlock()

	if len(guildFilters) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no filters.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Iterates through all the filters in memory and adds them to the filters string
	for _, filter := range guildFilters {
		filters += fmt.Sprintf("**%s**\n", filter.Filter)
	}
	filters = strings.TrimSuffix(filters, "\n")

	// Splits and sends message
	splitMessage := functionality.SplitLongMessage(filters)
	for i := 0; i < len(splitMessage); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot send filters message.")
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
}

// Adds a message requirement phrase to storage and memory
func addMessRequirementCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		channelID       string
		requirementType string
		phrase          string
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 4)

	if len(commandStrings) == 1 {
		message := fmt.Sprintf("Usage: `%vmrequire [channel]* [type]* [phrase]`\n\n[channel] is a ping or ID to the channel where the requirement will only be done.\n"+
			"[type] can either be soft or hard. Soft means a user must mention the phrase in their first message and is okay until someone else types a message. Hard means all messages must contain that phrase. Defaults to soft.\n"+
			"[phrase] is either regex expression (preferable) or just a simple string.\n\n"+
			"***** is optional.", guildSettings.Prefix)
		_, err := s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}
	// Resolves optional parameters based on commandStrings length
	if len(commandStrings) == 2 {
		phrase = commandStrings[1]
	} else if len(commandStrings) == 3 {
		channelID, _ = functionality.ChannelParser(s, commandStrings[1], m.GuildID)
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
		channelID, _ = functionality.ChannelParser(s, commandStrings[1], m.GuildID)
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

	// Removes arrows from emojis, mentions and channels
	phrase = strings.TrimPrefix(phrase, "<")
	phrase = strings.TrimSuffix(phrase, ">")

	// Writes the phrase to messrequirement.json and checks if the requirement was already in storage
	err := functionality.MessRequirementWrite(phrase, channelID, requirementType, m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%v` has been added to the message requirement list.", phrase))
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// Removes a message requirement from storage and memory
func removeMessRequirementCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		channelID string
		phrase    string
	)

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	messsageRequirementsLen := len(functionality.GuildMap[m.GuildID].MessageRequirements)
	functionality.Mutex.RUnlock()

	if messsageRequirementsLen == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no message requirements.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	commandStrings := strings.SplitN(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ", 3)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vunmrequire [channel]* [phrase]`\n\n[channel] is the channel for which that message requirement was set.\n"+
			"`[phrase]` is the phrase that was used when creating a message requirement.\n\n ***** are optional.", guildSettings.Prefix))
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Resolves optional parameter
	if len(commandStrings) == 3 {
		channelID, _ = functionality.ChannelParser(s, commandStrings[1], m.GuildID)
		if channelID == "" {
			phrase = commandStrings[1] + " " + commandStrings[2]
		} else {
			phrase = commandStrings[2]
		}
	} else {
		phrase = commandStrings[1]
	}

	// Removes arrows from emojis, mentions and channels
	phrase = strings.TrimPrefix(phrase, "<")
	phrase = strings.TrimSuffix(phrase, ">")

	// Removes the phrase from storage and memory
	err := functionality.MessRequirementRemove(phrase, channelID, m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("`%v` has been removed from the message requirement list.", phrase))
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

// Print message requirements from memory in chat
func viewMessRequirementCommand(s *discordgo.Session, m *discordgo.Message) {

	var mRequirements string

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	messsageRequirements := functionality.GuildMap[m.GuildID].MessageRequirements
	functionality.Mutex.RUnlock()

	if len(messsageRequirements) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no message requirements.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Iterates through all the message requirements in memory and adds them to the mRequirements string
	for _, requirement := range messsageRequirements {
		if requirement.Channel == "" {
			requirement.Channel = "All channels"
		}
		mRequirements += fmt.Sprintf("**%s** - **%s** - **%s**\n", requirement.Phrase, requirement.Channel, requirement.Type)
	}
	mRequirements = strings.TrimSuffix(mRequirements, "\n")

	// Splits and sends message
	splitMessage := functionality.SplitLongMessage(mRequirements)
	for i := 0; i < len(splitMessage); i++ {
		_, err := s.ChannelMessageSend(m.ChannelID, splitMessage[i])
		if err != nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Cannot send message requirements message.")
			if err != nil {
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
		}
	}
}

//// Removes user message if sent too quickly in succession
//func SpamFilter(s *discordgo.Session, m *discordgo.MessageCreate) {
//
//	// Checks if the bot had thrown an error before and stops it if so. Helps with massive backlog or delays but disables spam filter
//	if spamFilterIsBroken {
//		return
//	}
//	// Stops double event bug with empty content
//	if m.Content == "" {
//		return
//	}
//	// Checks if it's the bot that sent the message
//	if m.Author.ID == s.State.User.ID {
//		return
//	}
//
//	functionality.Mutex.Lock()
//	functionality.HandleNewGuild(s, m.GuildID)
//	functionality.Mutex.Unlock()
//
//	// Pulls info on message author
//	mem, err := s.State.Member(m.GuildID, m.Author.ID)
//	if err != nil {
//		mem, err = s.GuildMember(m.GuildID, m.Author.ID)
//		if err != nil {
//			return
//		}
//	}
//	// Checks if user is mod or bot before checking the message
//	functionality.Mutex.Lock()
//	if functionality.HasElevatedPermissions(s, mem.User.ID, m.GuildID) {
//		functionality.Mutex.Unlock()
//		return
//	}
//
//	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
//
//	// Counter for how many rapidly sent user messages a user has
//	if spamFilterMap[m.Author.ID] < 4 {
//		spamFilterMap[m.Author.ID]++
//		functionality.Mutex.Unlock()
//		return
//	}
//
//	// Stops filter if there is an unusual high size of messages to be deleted from a user (i.e. discord lags)
//	if spamFilterMap[m.Author.ID] > 15 {
//		if guildSettings.BotLog != nil {
//			if guildSettings.BotLog.ID != "" {
//				_, err = s.ChannelMessageSend(guildSettings.BotLog.ID, "Error: My spam filter has been disabled due to massive overflow of requests.")
//				if err != nil {
//					spamFilterIsBroken = true
//					functionality.Mutex.Unlock()
//					return
//				}
//			}
//		}
//		spamFilterIsBroken = true
//		functionality.Mutex.Unlock()
//		return
//	}
//	functionality.Mutex.Unlock()
//
//	// Deletes the message if over 4 rapidly sent messages
//	err = s.ChannelMessageDelete(m.ChannelID, m.ID)
//	if err != nil {
//		return
//	}
//}

//// Handles expiring user spam filter map counter
//func SpamFilterTimer(s *discordgo.Session, e *discordgo.Ready) {
//	for range time.NewTicker(4 * time.Second).C {
//		functionality.Mutex.Lock()
//		for userID := range spamFilterMap {
//			if spamFilterMap[userID] > 0 {
//				spamFilterMap[userID]--
//			}
//		}
//		functionality.Mutex.Unlock()
//	}
//}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    viewFiltersCommand,
		Trigger:    "filters",
		Aliases:    []string{"viewfilters", "viewfilter"},
		Desc:       "Prints all current filters",
		Permission: functionality.Mod,
		Module:     "filters",
	})
	functionality.Add(&functionality.Command{
		Execute:    addFilterCommand,
		Trigger:    "addfilter",
		Aliases:    []string{"filter", "setfilter"},
		Desc:       "Adds a phrase to the filters list. Works for reacts and emotes too. User regex for more complex filters",
		Permission: functionality.Mod,
		Module:     "filters",
	})
	functionality.Add(&functionality.Command{
		Execute:    removeFilterCommand,
		Trigger:    "removefilter",
		Aliases:    []string{"deletefilter", "unfilter"},
		Desc:       "Removes a phrase from the filters list",
		Permission: functionality.Mod,
		Module:     "filters",
	})
	functionality.Add(&functionality.Command{
		Execute:    viewMessRequirementCommand,
		Trigger:    "mrequirements",
		Aliases:    []string{"viewmrequirements", "showmrequirements", "messagerequirements", "messagereqirement", "viewmessrequirements", "messrequirements", "mrequirement", "messrequirement"},
		Desc:       "Prints all current message requirement filters",
		Permission: functionality.Mod,
		Module:     "filters",
	})
	functionality.Add(&functionality.Command{
		Execute:    addMessRequirementCommand,
		Trigger:    "mrequire",
		Aliases:    []string{"messrequire", "setmrequire", "setmessrequire", "setmessagerequire", "addmrequire", "messagerequire", "messrequire", "addmessrequire", "addmessagereqyure"},
		Desc:       "Adds a phrase to the message requirement list where it will remove messages that do not contain it",
		Permission: functionality.Mod,
		Module:     "filters",
	})
	functionality.Add(&functionality.Command{
		Execute:    removeMessRequirementCommand,
		Trigger:    "unmrequire",
		Aliases:    []string{"munrequire", "removemrequire", "removemrequirement", "deletemrequire", "deletemrequirement", "unmessrequire", "deletemessrequire", "removemessrequire"},
		Desc:       "Removes a phrase from the message requirement list",
		Permission: functionality.Mod,
		Module:     "filters",
	})
}
