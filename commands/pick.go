package commands

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Picks one item from a specified number of item.
func pickCommand(s *discordgo.Session, m *discordgo.Message) {

	var guildSettings = &functionality.GuildSettings{
		Prefix: ".",
	}

	if m.GuildID != "" {
		functionality.MapMutex.Lock()
		*guildSettings = functionality.GuildMap[m.GuildID].GetGuildSettings()
		functionality.MapMutex.Unlock()
	}

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vpick [item], [item]...`\n\nItem is anything that does not contain `,`\nUse `|` insead of `,` if you need a comma in the item", guildSettings.Prefix))
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Splits each item individually
	items := strings.Split(commandStrings[1], "|")
	if len(items) == 1 {
		items = strings.Split(commandStrings[1], ",")
	}

	// Trims trailing and leading whitespace from each item. Also removes items that are empty
	for i := len(items) - 1; i >= 0; i-- {
		items[i] = strings.TrimSpace(items[i])
		if items[i] == "" {
			items = append(items[:i], items[i+1:]...)
		}
	}

	// Check if after the split the item is still one
	if len(items) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Not enough items. Please add at least one more item.")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Picks a random item
	randomItemNum := rand.Intn(len(items))
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**Picked:** %v", items[randomItemNum]))
	if err != nil {
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute: pickCommand,
		Trigger: "pick",
		Aliases: []string{"pic", "pik", "p"},
		Desc:    "Picks a random item from a list of items",
		Module:  "normal",
		DMAble:  true,
	})
}
