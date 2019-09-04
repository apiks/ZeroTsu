package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/misc"
)

// Removes the previous x amount of messages in the channel it was used
func pruneCommand(s *discordgo.Session, m *discordgo.Message) {

	misc.MapMutex.Lock()
	guildPrefix := misc.GuildMap[m.GuildID].GuildConfig.Prefix
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	misc.MapMutex.Unlock()

	commandStrings := strings.Split(m.Content, " ")

	// Throw error not correct amoutn of parameters
	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Usage: `%vprune [amount]`\n\n[amount] is the number of messages to remove. Max is 5000.", guildPrefix))
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		return
	}
	// If an amount was specified then remove an x amount of messages
	if len(commandStrings) == 2 {
		amount, err := strconv.Atoi(commandStrings[1])
		if err != nil {
			if err.(*strconv.NumError).Err == strconv.ErrRange {
				misc.CommandErrorHandler(s, m, errors.New("Error: number out of range."), guildBotLog)
				return
			} else if err.(*strconv.NumError).Err == strconv.ErrSyntax {
				misc.CommandErrorHandler(s, m, errors.New("Error: not a valid number."), guildBotLog)
				return
			}
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}

		if amount > 5000 {
			misc.CommandErrorHandler(s, m, errors.New("Error: prune number is too large. Please use a smaller one."), guildBotLog)
			return
		}

		pruneMessages(s, m, amount, guildBotLog)
	}
}

// Removes the previous X amount of messages in X channel
func pruneMessages(s *discordgo.Session, m *discordgo.Message, amount int, guildBotLog string){
	var (
		deleteMessageIDs []string
		lastMessageID	 string
		successMessages  []string
	)

	// Add the command message
	deleteMessageIDs = append(deleteMessageIDs, m.ID)

	// Save current time
	now := time.Now()

	if amount <= 100 {
		successMess1, err := s.ChannelMessageSend(m.ChannelID, "Fetching messages to prune . . .")
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
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
				misc.CommandErrorHandler(s, m, err, guildBotLog)
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
			misc.CommandErrorHandler(s, m, err, guildBotLog)
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
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		return
	}

	if len(deleteMessageIDs) > 100 {
		successMess2, err := s.ChannelMessageSend(m.ChannelID, "Starting to prune messages. This might take a while . . .")
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		successMessages = append(successMessages, successMess2.ID)
	}

	// Deletes each 100 messages in the deleteMessageIDs in bulk
	if len(deleteMessageIDs) <= 100 {
		err := s.ChannelMessagesBulkDelete(m.ChannelID, deleteMessageIDs)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}
		successMess3, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Removed the past %v messages in this channel. Removing command messages in 2 seconds.", len(deleteMessageIDs)-1))
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
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
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}

		starti += 100
		if endi + 100 > len(deleteMessageIDs) {
			endi = len(deleteMessageIDs)
		} else {
			endi += 100
		}

		messagesLen -= 100
	}

	successMess3, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Success! Removed the past %v messages in this channel. Removing command messages in 2 seconds.", len(deleteMessageIDs)))
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Deletes success messages
	successMessages = append(successMessages, successMess3.ID)
	time.Sleep(2 * time.Second)
	_ = s.ChannelMessagesBulkDelete(successMess3.ChannelID, successMessages)
}

func init() {
	add(&command{
		execute:  pruneCommand,
		trigger:  "prune",
		aliases:  []string{"p", "prun", "pru", "purge"},
		desc:     "Prunes the previous x amount of messages in a channel. Works only for messages under 14 days old. Takes 5 seconds per 100 messages. MAX is 5000 which takes roughly four minutes.",
		elevated: true,
		category: "channel",
	})
}
