package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Picks one item from a specified number of item.
func pickCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		err           error
		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%spick [item], [item]...`\n\nItem is anything that does not contain `,`\nUse `|` insead of `,` if you need a comma in the item", guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
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
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Picks a random item
	randomItemNum := rand.Intn(len(items))
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**Picked:** %s", items[randomItemNum]))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

func init() {
	Add(&Command{
		Execute: pickCommand,
		Trigger: "pick",
		Aliases: []string{"pic", "pik", "p"},
		Desc:    "Picks a random item from a list of items",
		Module:  "normal",
		DMAble:  true,
	})
}
