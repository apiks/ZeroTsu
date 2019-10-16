package commands

import (
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
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
		functionality.MapMutex.Lock()
	}
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	if m.Author.ID != s.State.User.ID {
		functionality.MapMutex.Unlock()
	}

	commandStrings := strings.Split(strings.Replace(strings.ToLower(m.Content), "  ", " ", -1), " ")

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"sortcategory [category]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Fetches all channels from the server and puts it in deb
	deb, err := s.GuildChannels(m.GuildID)
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Invalid Module")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
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
	sort.Sort(functionality.SortChannelByAlphabet(categoryChannels))

	// Updates the alphabetically sorted channels' position
	for i := 0; i < len(categoryChannels); i++ {
		chaEdit.Position = categoryPosition + i + 1
		_, err = s.ChannelEditComplex(categoryChannels[i].ID, &chaEdit)
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Module `"+commandStrings[1]+"` sorted")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    sortCategoryCommand,
		Trigger:    "sortcategory",
		Desc:       "Sorts a category alphabetically",
		Permission: functionality.Mod,
		Module:     "misc",
	})
}
