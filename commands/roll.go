package commands

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/bwmarrin/discordgo"
)

// rollCommand rolls a number between 1 and a specified number (defaults to 100)
func rollCommand(max int) int {
	var result int

	if max <= 1 {
		result = rand.Intn(99) + 1
	} else if max > 1 {
		result = rand.Intn(max) + 1
	}

	return result
}

// rollCommandHandler rolls a number between 1 and a specified number (defaults to 100)
func rollCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		err           error
		guildSettings = entities.GuildSettings{Prefix: "."}
	)

	if m.GuildID != "" {
		guildSettings = db.GetGuildSettings(m.GuildID)
	}

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	// Rolls a number between 1 and 100 if only the command is used
	if len(commandStrings) == 1 {
		randomNum := rand.Intn(99) + 1
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**Rolled:** %d", randomNum))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Prints error if too many parameters
	if len(commandStrings) > 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sroll [number]`", guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if the specified number is actually a number
	num, err := strconv.Atoi(commandStrings[1])
	if err != nil {
		// Handles different error types
		if err.(*strconv.NumError).Err == strconv.ErrRange {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: That number is too large. Please try a smaller one.")
			if err != nil {
				common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
				return
			}
			return
		}
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: That is not a valid number.")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Makes sure it's not a number outside of boundaries
	if num < 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: Invalid number. Please use a positive number.")
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Rolls a specified number
	randomNum := rand.Intn(num) + 1
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**Rolled:** %v", randomNum))
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
}

func init() {
	Add(&Command{
		Execute: rollCommandHandler,
		Name:    "roll",
		Aliases: []string{"rol", "r"},
		Desc:    "Rolls a number from 1 to 100. Specify a positive number to change the range.",
		Module:  "normal",
		DMAble:  true,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "number",
				Description: "A positive number that specifies the range from 1 to the number.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var max int
			if i.ApplicationCommandData().Options != nil {
				for _, option := range i.ApplicationCommandData().Options {
					if option.Name == "number" {
						max = int(option.IntValue())
					}
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: strconv.Itoa(rollCommand(max)),
				},
			})
		},
	})
}
