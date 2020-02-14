package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"ZeroTsu/functionality"
)

// Removes the previous x amount of messages in the channel it was used
func pruneCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.Mutex.RLock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.Mutex.RUnlock()

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	// Throw error not correct amoutn of parameters
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%sprune [amount]`\n\n[amount] is the number of messages to remove. Max is 5000.", guildSettings.Prefix))
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}
	// If an amount was specified then remove an x amount of messages
	if len(commandStrings) == 2 {
		amount, err := strconv.Atoi(commandStrings[1])
		if err != nil {
			if err.(*strconv.NumError).Err == strconv.ErrRange {
				functionality.CommandErrorHandler(s, m, guildSettings.BotLog, errors.New("Error: number out of range."))
				return
			} else if err.(*strconv.NumError).Err == strconv.ErrSyntax {
				functionality.CommandErrorHandler(s, m, guildSettings.BotLog, errors.New("Error: not a valid number."))
				return
			}
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}

		if amount > 5000 {
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, errors.New("Error: prune number is too large. Please use a smaller one."))
			return
		}

		pruneMessages(s, m, amount, guildSettings.BotLog)
	}
}

// Removes the previous X amount of messages in X channel
func pruneMessages(s *discordgo.Session, m *discordgo.Message, amount int, guildBotLog *functionality.Cha) {
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
			functionality.CommandErrorHandler(s, m, guildBotLog, err)
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
				functionality.CommandErrorHandler(s, m, guildBotLog, err)
				return
			}
			for i := 0; i < len(messages); i++ {

				// Only save messages not older than 2 weeks
				timestamp, err := messages[i].Timestamp.Parse()
				if err != nil {
					continue
				}
				difference := now.Sub(timestamp)
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
			functionality.CommandErrorHandler(s, m, guildBotLog, err)
			return
		}
		for i := 0; i < len(messages); i++ {

			// Only save messages not older than 2 weeks
			timestamp, err := messages[i].Timestamp.Parse()
			if err != nil {
				continue
			}
			difference := now.Sub(timestamp)
			if difference.Hours() >= 336 {
				break OuterLoop
			}

			deleteMessageIDs = append(deleteMessageIDs, messages[i].ID)
		}
		amount -= 100
	}

	if len(deleteMessageIDs) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: the messages are more than 14 days old, cannot get the old messages or there are no other messages.")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildBotLog, err)
			return
		}
		return
	}

	if len(deleteMessageIDs) > 100 {
		successMess2, err := s.ChannelMessageSend(m.ChannelID, "Starting to prune messages. This might take a while . . .")
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildBotLog, err)
			return
		}
		successMessages = append(successMessages, successMess2.ID)
	}

	// Deletes each 100 messages in the deleteMessageIDs in bulk
	if len(deleteMessageIDs) <= 100 {
		err := s.ChannelMessagesBulkDelete(m.ChannelID, deleteMessageIDs)
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildBotLog, err)
			return
		}
		successMess3, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Removed the past %v messages in this channel. Removing command messages in 2 seconds.", len(deleteMessageIDs)-1))
		if err != nil {
			functionality.CommandErrorHandler(s, m, guildBotLog, err)
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
			functionality.CommandErrorHandler(s, m, guildBotLog, err)
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
		functionality.CommandErrorHandler(s, m, guildBotLog, err)
		return
	}

	// Deletes success messages
	successMessages = append(successMessages, successMess3.ID)
	time.Sleep(2 * time.Second)
	_ = s.ChannelMessagesBulkDelete(successMess3.ChannelID, successMessages)
}

func init() {
	functionality.Add(&functionality.Command{
		Execute:    pruneCommand,
		Trigger:    "prune",
		Aliases:    []string{"p", "prun", "pru", "purge"},
		Desc:       "Prunes the previous x amount of messages in a channel. Works only for messages under 14 days old. Takes 5 seconds per 100 messages. MAX is 5000 which takes roughly four minutes.",
		Permission: functionality.Mod,
		Module:     "channel",
	})
}
