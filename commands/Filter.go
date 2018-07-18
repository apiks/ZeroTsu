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

//Handles filter addition, removal and message check
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
	if ch.GuildID == config.ServerID {

		//Pulls info on message author
		mem, err := s.State.Member(config.ServerID, m.Author.ID)
		if err != nil {
			mem, err = s.GuildMember(config.ServerID, m.Author.ID)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}

		//Puts the command to lowercase
		messageLowercase := strings.ToLower(m.Content)

		//Checks if user has permissions and whether the BotPrefix was used
		if strings.HasPrefix(messageLowercase, config.BotPrefix) {
			if misc.HasPermissions(mem) {
				if strings.HasPrefix(messageLowercase, config.BotPrefix+"addfilter ") && (messageLowercase != (config.BotPrefix + "addfilter")) {

					if m.Author.ID == config.BotID {
						return
					}

					//Assigns the word to be filtered to the "word" variable
					word := strings.Replace(messageLowercase, config.BotPrefix+"addfilter ", "", -1)

					//Calls the function to write the new filter word to filters.json
					misc.FiltersWrite(word)

					if misc.FilterExists == false {

						//Prints success
						success := "`" + word + "` has been added to the filter list."
						_, err = s.ChannelMessageSend(m.ChannelID, success)
						if err != nil {

							fmt.Println("Error: ", err)
						}
					} else {

						//Prints failure
						failure := "`" + word + "` is already on the filter list."
						_, err = s.ChannelMessageSend(m.ChannelID, failure)
						if err != nil {

							fmt.Println("Error: ", err)
						}
					}

				} else if messageLowercase == config.BotPrefix+"filters" {
					if m.Author.ID == config.BotID {
						return
					}

					//Reads all the filters from filters.json
					misc.FiltersRead()

					//Creates a string variable to store the filters in for showing later
					var filters string

					//Iterates through all the filters if they exist and adds them to the filters string
					if len(misc.ReadFilters) != 0 {
						for i := 0; i < len(misc.ReadFilters); i++ {

							if filters == "" {

								filters = "`" + misc.ReadFilters[i].Filter + "`"
							} else {

								filters = filters + "\n `" + misc.ReadFilters[i].Filter + "`"
							}
						}
					}

					//If there are no filtered words give error, else print the filtered words.
					if len(misc.ReadFilters) == 0 {

						failure := "Error. There are no filters."
						_, err = s.ChannelMessageSend(m.ChannelID, failure)
						if err != nil {

							fmt.Println("Error: ", err)
						}
					} else {

						_, err = s.ChannelMessageSend(m.ChannelID, filters)
						if err != nil {

							fmt.Println("Error: ", err)
						}
					}

				} else if strings.HasPrefix(messageLowercase, config.BotPrefix+"removefilter ") && (messageLowercase != (config.BotPrefix + "removefilter")) {

					if m.Author.ID == config.BotID {
						return
					}

					//Reads all the filters from filters.json
					misc.FiltersRead()

					//Assigns the word to be filtered to the "word" variable
					word := strings.Replace(messageLowercase, config.BotPrefix+"removefilter ", "", -1)

					//Checks if there's any filters, else prints success.
					if len(misc.ReadFilters) == 0 {

						failure := "Error. There are no filters."

						_, err = s.ChannelMessageSend(m.ChannelID, failure)
						if err != nil {

							fmt.Println("Error: ", err)

						}
					} else {

						//Calls the function to remove the word from filters.json
						misc.FiltersRemove(word)

						//Prints success
						success := "`" + word + "` has been removed from the filter list."
						_, err = s.ChannelMessageSend(m.ChannelID, success)
						if err != nil {

							fmt.Println("Error: ", err)
						}
					}
				}
			}
		}

		//Checks if user is mod or bot before checking the message
		if misc.HasPermissions(mem) == false {
			if m.Author.ID == config.BotID {
				return
			}

			//Initializes a string in which if a word is removed it'll be stored for printing
			//Also initializes a bool which'll be used to measure against in printing
			var (
				removals      string
				badWordExists bool
			)

			//Reads all the filters from filters.json
			misc.FiltersRead()

			//Iterates through all the filters to see if the message contained a filtered word
			for i := 0; i < len(misc.ReadFilters); i++ {

				re := regexp.MustCompile(misc.ReadFilters[i].Filter)
				badWordCheck := re.FindAllString(messageLowercase, -1)

				if badWordCheck != nil {

					//Bool value for outside of loop to print the removals
					badWordExists = true

					//Deletes the message that was sent if it has a filtered word.
					s.ChannelMessageDelete(m.ChannelID, m.ID)

					//Stores the removals for printing
					if len(removals) == 0 {

						removals = badWordCheck[0]
					} else {

						removals = removals + ", " + badWordCheck[0]
					}
				}
			}

			if badWordExists == true {

				//Stores time of removal
				t := time.Now()
				z, _ := t.Zone()
				now := t.Format("2006-01-02 15:04:05") + " " + z

				//Sends embed mod message
				err := FilterEmbed(s, m, removals, now, m.ChannelID)
				if err != nil {
					l.Println(err)
				}

				//Assigns success print string for user
				success := "Your message `" + messageLowercase + "` was removed for using: _" + removals + "_ \n" +
					"Using such words makes me disappointed in you, darling."

				//Creates a DM connection and assigns it to dm
				dm, err := s.UserChannelCreate(m.Author.ID)
				if err != nil {
					l.Println("Error: ", err)
				}

				//Sends a message to that DM connection
				_, err = s.ChannelMessageSend(dm.ID, success)
				if err != nil {
					l.Println("Error: ", err)
				}
			}
		}
	}
}

// Filters reactions that contain a filtered phrase
func FilterReacts(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	//Reads all the filters from filters.json
	misc.FiltersRead()

	//Iterates through all the filters to see if the message contained a filtered word
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

			//Deletes the reaction that was sent if it has a filtered word
			err := s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.APIName(), r.UserID)
			if err != nil {

				fmt.Println("Error: ", err)
			}
		}
	}
}

func FilterEmbed(s *discordgo.Session, m *discordgo.MessageCreate, removals, now, channelID string) error {

	//Initializing needed variables for the embed
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
	embedFieldChannel.Value = chMentionID(channelID)

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

	//Send embed in bot-log channel
	_, err := s.ChannelMessageSendEmbed(config.BotLogID, &embedMess)
	return err
}
