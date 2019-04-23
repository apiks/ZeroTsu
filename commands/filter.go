package commands

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

var (
	spamFilterMap = make(map[string]int)
	spamFilterIsBroken = false
)

// Handles filter in an onMessage basis
func FilterHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	var (
		removals      	string
		badWordsSlice 	[]string
		badWordExists 	bool
		mLowercase		string
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				fmt.Println(rec)
			}
		}
	}()

	// Stops double event bug with empty content
	if m.Content == "" {
		return
	}
	// Checks if it's within the config server
	if m.GuildID != config.ServerID {
		return
	}
	// Checks if it's the bot that sent the message
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Pulls info on message author
	mem, err := s.State.Member(config.ServerID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, m.Author.ID)
		if err != nil {
			return
		}
	}
	// Checks if user is mod or bot before checking the message
	if misc.HasPermissions(mem) {
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
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
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
	_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message `%v` was removed for using: _%v_ \n\n" +
		"Using such words makes me disappointed in you, darling.", mLowercase, removals))
}

// Handles filter in an onEdit basis
func FilterEditHandler(s *discordgo.Session, m *discordgo.MessageUpdate) {

	var (
		removals      	string
		badWordsSlice 	[]string
		badWordExists 	bool
		mLowercase		string
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				fmt.Println(rec)
			}
		}
	}()

	// Stops double event bug with empty content
	if m.Content == "" {
		return
	}
	// Checks if it's within the config server
	if m.GuildID != config.ServerID {
		return
	}
	// Checks if it's the bot that sent the message
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Pulls info on message author
	mem, err := s.State.Member(config.ServerID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, m.Author.ID)
		if err != nil {
			return
		}
	}
	// Checks if user is mod or bot before checking the message
	if misc.HasPermissions(mem) {
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
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
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
	_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("Your message `%v` was removed for using: _%v_ \n\n" +
		"Using such words makes me disappointed in you, darling.", mLowercase, removals))
}

// Filters reactions that contain a filtered phrase
func FilterReactsHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	var badReactExists	bool

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string))
			if err != nil {
				fmt.Println(rec)
			}
		}
	}()

	// Checks if it's within the config server
	if r.GuildID != config.ServerID {
		return
	}
	// Checks if it's the bot that sent the message
	if r.UserID == s.State.User.ID {
		return
	}
	// Pulls info on message author
	mem, err := s.State.Member(config.ServerID, r.UserID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, r.UserID)
		if err != nil {
			return
		}
	}
	// Checks if user is mod or bot before checking the message
	if misc.HasPermissions(mem) {
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
		_, _ = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
	}
}

// Checks if the message is supposed to be filtered
func isFiltered(s *discordgo.Session, m *discordgo.Message) (bool, []string){

	var (
		mLowercase				string
		mentions				string
		userID					string

		badPhraseSlice         []string
		badPhraseCheckMentions []string
		badPhraseCheck         []string
		mentionCheck           []string
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
				if len(misc.MemberInfoMap) != 0 {
					if _, ok := misc.MemberInfoMap[userID]; ok {
						mentions += " " + strings.ToLower(misc.MemberInfoMap[userID].Nickname)
						misc.MapMutex.Unlock()
						continue
					}
				}
				misc.MapMutex.Unlock()

				// If user wasn't found in memberInfo with that username+discrim combo then fetch manually from Discord
				user, err := s.State.Member(config.ServerID, userID)
				if err != nil {
					user, _ = s.GuildMember(config.ServerID, userID)
				}
				if user != nil {
					mentions += " " + strings.ToLower(user.Nick)
				}
			}
		}
	}

	// Iterates through all the filters to see if the message contained a filtered phrase
	misc.MapMutex.Lock()
	for _, filter := range misc.ReadFilters {

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

	return false, nil
}

// Checks if the React is supposed to be filtered
func isFilteredReact(s *discordgo.Session, r *discordgo.MessageReactionAdd) bool {

	var reactName	string

	// Iterates through all the filters to see if the react contained a filtered phrase
	misc.MapMutex.Lock()
	for _, filter := range misc.ReadFilters {

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

// Adds a filter to storage and memory
func addFilterCommand(s *discordgo.Session, m *discordgo.Message) {

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "filter [phrase]`")
		if err != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Writes to filters.json
	filterExists, err := misc.FiltersWrite(commandStrings[1])
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	if !filterExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "`" + commandStrings[1] + "` has been added to the filter list.")
		if err != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: `" + commandStrings[1] + "` is already on the filter list.")
		if err != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {
				return
			}
			return
		}
	}
}

// Removes a filter from storage and memory
func removeFilterCommand(s *discordgo.Session, m *discordgo.Message) {

	if len(misc.ReadFilters) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no filters.")
		if err != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.SplitN(messageLowercase, " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "unfilter [phrase]`")
		if err != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Removes phrase from storage and memory
	filterExists, err := misc.FiltersRemove(commandStrings[1])
	if err != nil {
		misc.CommandErrorHandler(s, m, err)
		return
	}

	if filterExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "`" + commandStrings[1] + "` has been removed from the filter list.")
		if err != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Error: `" + commandStrings[1] + "` is not in the filter list.")
	if err != nil {
		_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Print filters from memory in chat
func viewFiltersCommand(s *discordgo.Session, m *discordgo.Message) {

	// Creates a string variable to store the filters in for showing later
	var filters string

	if len(misc.ReadFilters) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no filters.")
		if err != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Iterates through all the filters if they exist and adds them to the filters string
	for i := 0; i < len(misc.ReadFilters); i++ {
		if filters == "" {
			filters = "`" + misc.ReadFilters[i].Filter + "`"
		} else {
			filters = filters + "\n `" + misc.ReadFilters[i].Filter + "`"
		}
	}

	_, err := s.ChannelMessageSend(m.ChannelID, filters)
	if err != nil {
		_, err := s.ChannelMessageSend(config.BotLogID, err.Error())
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
	)

	// Checks if the message contains a mention and finds the actual name instead of ID
	content := m.Content
	content = misc.MentionParser(s, content)

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
	embedFieldFilter.Value = "**__" + removals + "__**"
	embedFieldMessage.Value = "`" + content + "`"
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
	_, err := s.ChannelMessageSendEmbed(config.BotLogID, &embedMess)
	return err
}

// Removes user message if sent too quickly in succession
func SpamFilter(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Checks if the bot had thrown an error before and stops it if so. Helps with massive backlog or delays but disables spam filter
	if spamFilterIsBroken {
		return
	}
	// Checks if it's within the /r/anime server
	s.RWMutex.Lock()
	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		ch, err = s.Channel(m.ChannelID)
		if err != nil {
			s.RWMutex.Unlock()
			return
		}
		s.RWMutex.Unlock()
		return
	}
	if ch.GuildID != config.ServerID {
		s.RWMutex.Unlock()
		return
	}
	// Checks if it's the bot that sent the message
	if m.Author.ID == s.State.User.ID {
		s.RWMutex.Unlock()
		return
	}
	s.RWMutex.Unlock()
	// Pulls info on message author
	mem, err := s.State.Member(config.ServerID, m.Author.ID)
	if err != nil {
		mem, err = s.GuildMember(config.ServerID, m.Author.ID)
		if err != nil {
			return
		}
	}
	// Checks if user is mod or bot before checking the message
	s.RWMutex.RLock()
	if misc.HasPermissions(mem) {
		s.RWMutex.RUnlock()
		return
	}
	s.RWMutex.RUnlock()

	// Removes message if there were over 4 rapidly sent messages
	misc.MapMutex.Lock()
	if spamFilterMap[m.Author.ID] < 4 {
		spamFilterMap[m.Author.ID]++
	} else {
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil && spamFilterMap[m.Author.ID] > 15 {
			_, err := s.ChannelMessageSend(config.BotLogID, "Error: Spam filter has been disabled due to massive overflow of requests.\n"+
				err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			spamFilterIsBroken = true
			misc.MapMutex.Unlock()
			return
		}
	}
	misc.MapMutex.Unlock()
}

// Handles expiring user spam map
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
}