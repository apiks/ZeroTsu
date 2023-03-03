package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// pruneCommand removes the previous x amount of messages in the channel it was used
func pruneCommand(s *discordgo.Session, amount int, targetChannelID string) error {
	if amount < 1 {
		return errors.New("Error: Invalid amount given. Minimum is 1.")
	}
	if amount > 5000 {
		return errors.New("Error: Amount is too large. Maximum is 5000.")
	}

	return pruneMessages(s, amount, targetChannelID)
}

// pruneMessages removes the previous X amount of messages in X channel
func pruneMessages(s *discordgo.Session, amount int, targetChannelID string) error {
	var (
		lastMessage      *discordgo.Message
		deleteMessageIDs []string
		lastMessageID    string
	)

	// Save current time
	now := time.Now()

	// Find a starting point
	lastMessages, err := s.ChannelMessages(targetChannelID, 2, "", "", "")
	if err != nil {
		return err
	}
	if len(lastMessages) == 0 {
		return errors.New("No valid messages could be found to delete.")
	} else if len(lastMessages) == 1 {
		lastMessage = lastMessages[0]
	} else if len(lastMessages) == 2 {
		lastMessage = lastMessages[1]
	}
	deleteMessageIDs = append(deleteMessageIDs, lastMessage.ID)

	// Keep iterating until amount is zero
OuterLoop:
	for amount > 0 {

		// Reset and save new last message ID saved in the slice
		lastMessageID = deleteMessageIDs[len(deleteMessageIDs)-1]

		// If the amount is under or equal to 100, just fetch that amount of messages and exit loop, otherwise fetch max (100) and loop
		if amount <= 100 {
			messages, err := s.ChannelMessages(targetChannelID, amount, lastMessageID, "", "")
			if err != nil {
				return err
			}
			for i := 0; i < len(messages); i++ {

				// Only save messages not older than 2 weeks
				difference := now.Sub(messages[i].Timestamp)
				if difference.Hours() >= 336 {
					break OuterLoop
				}

				deleteMessageIDs = append(deleteMessageIDs, messages[i].ID)
			}
			break
		}

		// If the amount is greater than 100 then fetch those messages and reduce amount
		messages, err := s.ChannelMessages(targetChannelID, 100, lastMessageID, "", "")
		if err != nil {
			return err
		}
		for i := 0; i < len(messages); i++ {

			// Only save messages not older than 2 weeks
			difference := now.Sub(messages[i].Timestamp)
			if difference.Hours() >= 336 {
				break OuterLoop
			}

			deleteMessageIDs = append(deleteMessageIDs, messages[i].ID)
		}
		amount -= 100
	}

	if len(deleteMessageIDs) == 1 {
		return errors.New("Error: The messages I tried deleting are either more than 14 days old, I cannot fetch them, or there are no other valid messages to prune.")
	}

	// Deletes each 100 messages in the deleteMessageIDs in bulk
	if len(deleteMessageIDs) <= 100 {
		err := s.ChannelMessagesBulkDelete(targetChannelID, deleteMessageIDs)
		if err != nil {
			return err
		}
		return nil
	}

	messagesLen := len(deleteMessageIDs)
	starti := 0
	endi := 100
	for messagesLen > 0 {

		err := s.ChannelMessagesBulkDelete(targetChannelID, deleteMessageIDs[starti:endi])
		if err != nil {
			return err
		}

		starti += 100
		if endi+100 > len(deleteMessageIDs) {
			endi = len(deleteMessageIDs)
		} else {
			endi += 100
		}

		messagesLen -= 100
	}

	return nil
}

// pruneCommandHandler removes the previous x amount of messages in the channel it was used
func pruneCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	// Throw error not correct amount of parameters
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sprune [amount]`\n\n[amount] is the number of messages to remove. Max is 5000.", guildSettings.GetPrefix()))
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}
	// If an amount was specified then remove an x amount of messages
	if len(commandStrings) == 2 {
		amount, err := strconv.Atoi(commandStrings[1])
		if err != nil {
			if err.(*strconv.NumError).Err == strconv.ErrRange {
				common.CommandErrorHandler(s, m, guildSettings.BotLog, errors.New("Error: number out of range."))
				return
			} else if err.(*strconv.NumError).Err == strconv.ErrSyntax {
				common.CommandErrorHandler(s, m, guildSettings.BotLog, errors.New("Error: not a valid number."))
				return
			}
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}

		if amount > 5000 {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, errors.New("Error: prune number is too large. Please use a smaller one."))
			return
		}

		pruneMessagesHandler(s, m, amount, guildSettings.BotLog)
	}
}

// pruneMessagesHandler removes the previous X amount of messages in X channel
func pruneMessagesHandler(s *discordgo.Session, m *discordgo.Message, amount int, guildBotLog entities.Cha) {
	var (
		deleteMessageIDs []string
		lastMessageID    string
		successMessages  []string
	)

	// Add the command message
	deleteMessageIDs = append(deleteMessageIDs, m.ID)

	// Save current time
	now := time.Now()

	if amount <= 100 {
		successMess1, err := s.ChannelMessageSend(m.ChannelID, "Fetching messages to prune . . .")
		if err != nil {
			common.CommandErrorHandler(s, m, guildBotLog, err)
			return
		}
		successMessages = append(successMessages, successMess1.ID)
	}

	// Keep iterating until amount is zero
OuterLoop:
	for amount > 0 {

		// Reset and save new last message ID saved in the slice
		lastMessageID = deleteMessageIDs[len(deleteMessageIDs)-1]

		// If the amount is under or equal to 100, just fetch that amount of messages and exit loop, otherwise fetch max (100) and loop
		if amount <= 100 {
			messages, err := s.ChannelMessages(m.ChannelID, amount, lastMessageID, "", "")
			if err != nil {
				common.CommandErrorHandler(s, m, guildBotLog, err)
				return
			}
			for i := 0; i < len(messages); i++ {

				// Only save messages not older than 2 weeks
				difference := now.Sub(messages[i].Timestamp)
				if difference.Hours() >= 336 {
					break OuterLoop
				}

				deleteMessageIDs = append(deleteMessageIDs, messages[i].ID)
			}
			break
		}

		// If the amount is greater than 100 then fetch those messages and reduce amount
		messages, err := s.ChannelMessages(m.ChannelID, 100, lastMessageID, "", "")
		if err != nil {
			common.CommandErrorHandler(s, m, guildBotLog, err)
			return
		}
		for i := 0; i < len(messages); i++ {

			// Only save messages not older than 2 weeks
			difference := now.Sub(messages[i].Timestamp)
			if difference.Hours() >= 336 {
				break OuterLoop
			}

			deleteMessageIDs = append(deleteMessageIDs, messages[i].ID)
		}
		amount -= 100
	}

	if len(deleteMessageIDs) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: The messages I tried deleting are either more than 14 days old, I cannot fetch them, or there are no other valid messages to prune.")
		if err != nil {
			common.CommandErrorHandler(s, m, guildBotLog, err)
			return
		}
		return
	}

	if len(deleteMessageIDs) > 100 {
		successMess2, err := s.ChannelMessageSend(m.ChannelID, "Starting to prune messages. This might take a while . . .")
		if err != nil {
			common.CommandErrorHandler(s, m, guildBotLog, err)
			return
		}
		successMessages = append(successMessages, successMess2.ID)
	}

	// Deletes each 100 messages in the deleteMessageIDs in bulk
	if len(deleteMessageIDs) <= 100 {
		err := s.ChannelMessagesBulkDelete(m.ChannelID, deleteMessageIDs)
		if err != nil {
			common.CommandErrorHandler(s, m, guildBotLog, err)
			return
		}
		successMess3, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Removed the past %v messages in this channel. Removing command messages in 2 seconds.", len(deleteMessageIDs)-1))
		if err != nil {
			common.CommandErrorHandler(s, m, guildBotLog, err)
			return
		}

		successMessages = append(successMessages, successMess3.ID)
		time.Sleep(2 * time.Second)
		_ = s.ChannelMessagesBulkDelete(successMess3.ChannelID, successMessages)
		return
	}

	messagesLen := len(deleteMessageIDs)
	starti := 0
	endi := 100
	for messagesLen > 0 {

		err := s.ChannelMessagesBulkDelete(m.ChannelID, deleteMessageIDs[starti:endi])
		if err != nil {
			common.CommandErrorHandler(s, m, guildBotLog, err)
			return
		}

		starti += 100
		if endi+100 > len(deleteMessageIDs) {
			endi = len(deleteMessageIDs)
		} else {
			endi += 100
		}

		messagesLen -= 100
	}

	successMess3, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Removed the past %d messages in this channel. Removing command messages in 2 seconds.", len(deleteMessageIDs)))
	if err != nil {
		common.CommandErrorHandler(s, m, guildBotLog, err)
		return
	}

	// Deletes success messages
	successMessages = append(successMessages, successMess3.ID)
	time.Sleep(2 * time.Second)
	_ = s.ChannelMessagesBulkDelete(successMess3.ChannelID, successMessages)
}

func init() {
	Add(&Command{
		Execute:    pruneCommandHandler,
		Name:       "prune",
		Aliases:    []string{"p", "prun", "pru", "purge"},
		Desc:       "Prunes the previous x amount of messages. Messages must not be older than 14 days. MAX is 5000.",
		Permission: functionality.Mod,
		Module:     "channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "number",
				Description: "A positive number that specifies the number of messages to delete. Up to 5000.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel in which to delete messages.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Please wait...",
				},
			})

			err := VerifySlashCommand(s, "prune", i)
			if err != nil {
				errStr := err.Error()
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			var (
				response        = "Fetching messages and beginning pruning. This might take a while. It may take up to a minute for the change to be reflected afterwards."
				amount          int
				targetChannelID = i.ChannelID
			)
			if i.ApplicationCommandData().Options == nil {
				return
			}

			if i.ApplicationCommandData().Options != nil {
				for _, option := range i.ApplicationCommandData().Options {
					if option.Name == "number" {
						amount = int(option.IntValue())
					} else if option.Name == "channel" {
						targetChannelID = option.ChannelValue(s).ID
					}
				}
			}

			if amount < 1 {
				response = "Error: Invalid amount given. Minimum is 1."
			}
			if amount > 5000 {
				response = "Error: Amount is too large. Maximum is 5000."
			}

			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &response,
			})

			err = pruneCommand(s, amount, targetChannelID)
			if err != nil {
				errStr := err.Error()
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			resp := fmt.Sprintf("Success! Removed the past %d valid messages in this channel. Deleting the command message in 3 seconds...", amount)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &resp,
			})
			time.Sleep(3 * time.Second)
			s.InteractionResponseDelete(i.Interaction)
		},
	})
}
