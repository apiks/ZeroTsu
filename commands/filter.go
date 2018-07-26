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

// Handles filter in an onMessage basis
func FilterHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Checks if it's within the /r/anime server
	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		ch, err = s.Channel(m.ChannelID)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
	if ch.GuildID != config.ServerID {

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
			fmt.Println(err.Error())
			return
		}
	}
	// Checks if user is mod or bot before checking the message
	if misc.HasPermissions(mem) == true {

		return
	}

	var (
		removals      string
		badWordsSlice []string
		badWordExists bool
	)

	// Puts the message to lowercase
	messageLowercase := strings.ToLower(m.Content)

	// Checks if message should be filtered
	badWordExists, badWordsSlice = isFiltered(m.Message)

	// If function returns true handle the filtered message
	if badWordExists {

		//Deletes the message that was sent if it has a filtered word.
		err = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {

			fmt.Println("Error:", err)
		}

		// Iterates through all the bad words
		for i := 0; i < len(badWordsSlice); i++ {
			// Stores the removals for printing
			if len(removals) == 0 {

				removals = badWordsSlice[0]
			} else {

				removals = removals + ", " + badWordsSlice[i]
			}
		}

		// Stores time of removal
		t := time.Now()
		z, _ := t.Zone()
		now := t.Format("2006-01-02 15:04:05") + " " + z

		// Sends embed mod message
		err := FilterEmbed(s, m.Message, removals, now, m.ChannelID)
		if err != nil {
			l.Println(err)
		}

		// Sends message to user's DMs
		dm, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			l.Println("Error:", err)
		}
		_, err = s.ChannelMessageSend(dm.ID, "Your message `" + messageLowercase + "` was removed for using: _" + removals + "_ \n" +
			"Using such words makes me disappointed in you, darling.")
		if err != nil {
			l.Println("Error:", err)
		}
	}
}

// Filters reactions that contain a filtered phrase
func FilterReactsHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	// Iterates through all the filters to see if the message contained a filtered word
	for i := 0; i < len(misc.ReadFilters); i++ {

		// Assigns the filter to a react variable so it can be changed to normal API mode name
		reactName := misc.ReadFilters[i].Filter

		// Trims the fluff from a reaction so it can measured against the API version below
		if strings.Contains(reactName, "<:") {

			reactName = strings.Replace(reactName, "<:", "", -1)
			reactName = strings.TrimSuffix(reactName, ">")

		} else if strings.Contains(reactName, "<a:") {

			reactName = strings.Replace(reactName, "<a:", "", -1)
			reactName = strings.TrimSuffix(reactName, ">")
		}

		re := regexp.MustCompile(reactName)
		badWordCheck := re.FindAllString(r.Emoji.APIName(), -1)

		if badWordCheck != nil {

			// Deletes the reaction that was sent if it has a filtered word
			err := s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.APIName(), r.UserID)
			if err != nil {

				fmt.Println("Error: ", err)
			}
		}
	}
}

// Checks if the message is supposed to be filtered
func isFiltered(m *discordgo.Message) (bool, []string){

	var (
		filtered bool
		badWordsSlice []string
	)

	// Puts the command to lowercase
	messageLowercase := strings.ToLower(m.Content)

	// Iterates through all the filters to see if the message contained a filtered word
	for i := 0; i < len(misc.ReadFilters); i++ {

		re := regexp.MustCompile(misc.ReadFilters[i].Filter)
		badWordCheck := re.FindAllString(messageLowercase, -1)

		if badWordCheck != nil {

			badWordsSlice = append(badWordsSlice, badWordCheck[0])
			filtered = true
		}
	}

	if filtered == true {

		return true, badWordsSlice
	} else {

		return false, nil
	}
}

// Adds a filter to storage and memory
func addFilterCommand(s *discordgo.Session, m *discordgo.Message) {

	if len(m.Content) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: not enough parameters")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}

	// Puts the command to lowercase
	messageLowercase := strings.ToLower(m.Content)

	// Parses the filtered phrase
	phrase := strings.Replace(messageLowercase, config.BotPrefix+"addfilter ", "", -1)

	// Writes to filters.json
	filterExists := misc.FiltersWrite(phrase)

	if filterExists == false {
		_, err := s.ChannelMessageSend(m.ChannelID, "`" + phrase + "` has been added to the filter list.")
		if err != nil {

			fmt.Println("Error:", err)
		}
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: `" + phrase + "` is already on the filter list.")
		if err != nil {

			fmt.Println("Error:", err)
		}
	}
}

// Removes a filter from storage and memory
func removeFilterCommand(s *discordgo.Session, m *discordgo.Message) {

	if len(m.Content) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: not enough parameters")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}
	if len(misc.ReadFilters) == 0 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no filters.")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}

	// Puts the command to lowercase
	messageLowercase := strings.ToLower(m.Content)

	// Parses the filtered phrase
	phrase := strings.Replace(messageLowercase, config.BotPrefix+"removefilter ", "", -1)

	// Removes phrase from storage and memory
	filterExists := misc.FiltersRemove(phrase)

	if filterExists == true {

		_, err := s.ChannelMessageSend(m.ChannelID, "`" + phrase + "` has been removed from the filter list.")
		if err != nil {

			fmt.Println("Error:", err)
		}
	} else {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: `" + phrase + "` is not in the filter list.")
		if err != nil {

			fmt.Println("Error:", err)
		}
	}
}

// Print filters from memory in chat
func viewFiltersCommand(s *discordgo.Session, m *discordgo.Message) {

	// Creates a string variable to store the filters in for showing later
	var filters string

	if len(misc.ReadFilters) == 0 {

		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no filters.")
		if err != nil {

			fmt.Println("Error:", err)
		}
		return
	}

	// Iterates through all the filters if they exist and adds them to the filters string
	if len(misc.ReadFilters) != 0 {
		for i := 0; i < len(misc.ReadFilters); i++ {

			if filters == "" {

				filters = "`" + misc.ReadFilters[i].Filter + "`"
			} else {

				filters = filters + "\n `" + misc.ReadFilters[i].Filter + "`"
			}
		}
	}

	_, err := s.ChannelMessageSend(m.ChannelID, filters)
	if err != nil {

		fmt.Println("Error:", err)
	}
}

func FilterEmbed(s *discordgo.Session, m *discordgo.Message, removals, now, channelID string) error {

	var (
		embedMess      discordgo.MessageEmbed
		embedThumbnail discordgo.MessageEmbedThumbnail

		//Embed slice and its fields
		embedField        []*discordgo.MessageEmbedField
		embedFieldFilter  discordgo.MessageEmbedField
		embedFieldMessage discordgo.MessageEmbedField
		embedFieldDate    discordgo.MessageEmbedField
		embedFieldChannel discordgo.MessageEmbedField
	)

	//Saves user avatar as thumbnail
	embedThumbnail.URL = m.Author.AvatarURL("128")

	//Sets field titles
	embedFieldFilter.Name = "Filtered:"
	embedFieldMessage.Name = "Message:"
	embedFieldDate.Name = "Date:"
	embedFieldChannel.Name = "Channel:"

	//Sets field content
	embedFieldFilter.Value = "**__" + removals + "__**"
	embedFieldMessage.Value = "`" + m.Content + "`"
	embedFieldDate.Value = now
	embedFieldChannel.Value = misc.ChMentionID(channelID)

	//Sets field inline
	embedFieldFilter.Inline = true
	embedFieldDate.Inline = true
	embedFieldChannel.Inline = true

	//Adds the two fields to embedField slice (because embedMess.Fields requires slice input)
	embedField = append(embedField, &embedFieldFilter)
	embedField = append(embedField, &embedFieldDate)
	embedField = append(embedField, &embedFieldChannel)
	embedField = append(embedField, &embedFieldMessage)

	//Sets embed title and its description (which it uses the same way as a field)
	embedMess.Title = "User:"
	embedMess.Description = m.Author.Mention()

	//Adds user thumbnail and the two other fields as well
	embedMess.Thumbnail = &embedThumbnail
	embedMess.Fields = embedField

	_, err := s.ChannelMessageSendEmbed(config.BotLogID, &embedMess)
	return err
}

// Adds filter commands to the commandHandler
func init() {
	add(&command{
		execute:  viewFiltersCommand,
		trigger:  "filters",
		aliases:  []string{"filter"},
		desc:     "Prints all current filters.",
		elevated: true,
	})
	add(&command{
		execute:  addFilterCommand,
		trigger:  "addfilter",
		desc:     "Adds a string to the filters list.",
		elevated: true,
	})
	add(&command{
		execute:  removeFilterCommand,
		trigger:  "removefilter",
		desc:     "Removes a string from the filters list.",
		elevated: true,
	})
}
