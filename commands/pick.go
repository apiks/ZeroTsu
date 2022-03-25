package commands

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/bwmarrin/discordgo"
)

// pickCommand picks one item from a specified number of item.
func pickCommand(items string) []string {
	var messages []string

	// Splits each item individually
	itemsSplit := strings.Split(items, "|")
	if len(itemsSplit) == 1 {
		itemsSplit = strings.Split(items, ",")
		if len(itemsSplit) == 1 {
			itemsSplit = strings.Split(items, "‚")
		}
	}

	// Trims trailing and leading whitespace from each item. Also removes items that are empty
	for i := len(itemsSplit) - 1; i >= 0; i-- {
		itemsSplit[i] = strings.TrimSpace(itemsSplit[i])
		if itemsSplit[i] == "" {
			itemsSplit = append(itemsSplit[:i], itemsSplit[i+1:]...)
		}
	}

	if len(itemsSplit) == 1 {
		return []string{"Error: At least 2 items required."}
	}

	// Picks a random item
	message := fmt.Sprintf("**Picked:** %s", itemsSplit[rand.Intn(len(itemsSplit))])

	// Splits the message if it's too big into multiple ones
	if len(message) > 1900 {
		messages = common.SplitLongMessage(message)
	}

	if messages == nil {
		return []string{message}
	}

	return messages
}

// pickCommandHandler picks one item from a specified number of item.
func pickCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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
		if len(items) == 1 {
			items = strings.Split(commandStrings[1], "‚")
		}
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
		Execute: pickCommandHandler,
		Name:    "pick",
		Aliases: []string{"pic", "pik", "p"},
		Desc:    "Picks a random item from a list of items.",
		Module:  "normal",
		DMAble:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "items",
				Description: "Items to select from, separate using a | or a comma (,). Minimum items required is 2.",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.ApplicationCommandData().Options == nil {
				return
			}

			items := ""
			if i.ApplicationCommandData().Options != nil {
				for _, option := range i.ApplicationCommandData().Options {
					if option.Name == "items" {
						items = option.StringValue()
					}
				}
			}

			messages := pickCommand(items)
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

					s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
						Content: message,
					})
				}
			}
		},
	})
}
