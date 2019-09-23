package commands

import (
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Sorts all channels in a given category alphabetically
func sortCategoryCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		categoryID       string
		categoryPosition int
		categoryChannels []*discordgo.Channel
		chaEdit          discordgo.ChannelEdit
	)

	if m.Author.ID != s.State.User.ID {
		misc.MapMutex.Lock()
	}
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	if m.Author.ID != s.State.User.ID {
		misc.MapMutex.Unlock()
	}

	messageLowercase := strings.ToLower(m.Content)
	commandStrings := strings.Split(messageLowercase, " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildPrefix+"sortcategory [category]`")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Fetches all channels from the server and puts it in deb
	deb, err := s.GuildChannels(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	for i := 0; i < len(deb); i++ {
		// Puts channel name to lowercase
		nameLowercase := strings.ToLower(deb[i].Name)

		// Compares if the categoryString is either a valid category name or ID
		if nameLowercase == commandStrings[1] || deb[i].ID == commandStrings[1] {
			if deb[i].Type == discordgo.ChannelTypeGuildCategory {
				categoryID = deb[i].ID
				categoryPosition = deb[i].Position
			}
		}
	}

	// Checks if category exists
	if categoryID == "" {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Invalid Category")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Puts all channels under a category in categoryChannels slice
	for i := 0; i < len(deb); i++ {
		if deb[i].ParentID == categoryID {
			categoryChannels = append(categoryChannels, deb[i])
		}
	}

	// Sorts the categoryChannels slice alphabetically
	sort.Sort(misc.SortChannelByAlphabet(categoryChannels))

	// Updates the alphabetically sorted channels' position
	for i := 0; i < len(categoryChannels); i++ {
		chaEdit.Position = categoryPosition + i + 1
		_, err = s.ChannelEditComplex(categoryChannels[i].ID, &chaEdit)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Category `"+commandStrings[1]+"` sorted")
	if err != nil {
		_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

func init() {
	add(&command{
		execute:  sortCategoryCommand,
		trigger:  "sortcategory",
		desc:     "Sorts a category alphabetically",
		elevated: true,
		category: "misc",
	})
}
