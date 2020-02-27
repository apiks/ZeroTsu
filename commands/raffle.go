package commands

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"log"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Assigns a user to participate in a raffle
func raffleParticipateCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		raffleExists bool
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildRaffles := db.GetGuildRaffles(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"jraffle [raffle name]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if such a raffle exists and adds the user ID to it if so
	for _, raffle := range guildRaffles {
		if raffle == nil {
			continue
		}

		if raffle.GetName() == strings.ToLower(commandStrings[1]) {
			raffleExists = true

			// Checks if the user already joined that raffle
			for _, ID := range raffle.GetParticipantIDs() {
				if ID == m.Author.ID {
					_, err := s.ChannelMessageSend(m.ChannelID, "You've already joined that raffle!")
					if err != nil {
						common.LogError(s, guildSettings.BotLog, err)
						return
					}
					return
				}
			}

			// Adds user ID to the raffle list
			db.SetGuildRaffleParticipant(m.GuildID, m.Author.ID, raffle)
			break
		}
	}

	if !raffleExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such raffle exists.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Success! You have entered raffle `"+commandStrings[1]+"`")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a user from a raffle
func raffleLeaveCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		raffleExists bool
		userInRaffle bool
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildRaffles := db.GetGuildRaffles(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"lraffle [raffle name]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if such a raffle exists and removes the user ID from it if so
	for _, raffle := range guildRaffles {
		if raffle == nil {
			continue
		}

		if raffle.GetName() == strings.ToLower(commandStrings[1]) {
			raffleExists = true

			// Checks if the user already joined that raffle and removes him if so
			for _, ID := range raffle.GetParticipantIDs() {
				if ID == m.Author.ID {
					userInRaffle = true

					// Leaves the raffle
					db.SetGuildRaffleParticipant(m.GuildID, m.Author.ID, raffle, true)
					break
				}
			}
			if !userInRaffle {
				_, err := s.ChannelMessageSend(m.ChannelID, "You're not in that raffle!")
				if err != nil {
					common.LogError(s, guildSettings.BotLog, err)
					return
				}
				return
			}
			break
		}
	}

	if !raffleExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such raffle exists.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Success! You have left raffle `"+commandStrings[1]+"`")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Enters a user in a raffle if they react
func RaffleReactJoin(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in RaffleReactJoin")
		}
	}()

	if r.GuildID == "" {
		return
	}

	entities.HandleNewGuild(r.GuildID)

	guildRaffles := db.GetGuildRaffles(r.GuildID)

	// Checks if it's the slot machine emoji or the bot itself
	if r.Emoji.APIName() != "🎰" {
		return
	}
	if r.UserID == s.State.User.ID {
		return
	}

	// Checks if that message has a raffle react set for it
	for _, raffle := range guildRaffles {
		if raffle == nil {
			continue
		}

		if raffle.GetReactMessageID() == r.MessageID {
			db.SetGuildRaffleParticipant(r.GuildID, r.UserID, raffle)
			return
		}
	}
}

// Removes a user from a raffle if they unreact
func RaffleReactLeave(s *discordgo.Session, r *discordgo.MessageReactionRemove) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in RaffleReactLeave")
		}
	}()

	if r.GuildID == "" {
		return
	}

	entities.HandleNewGuild(r.GuildID)

	guildRaffles := db.GetGuildRaffles(r.GuildID)

	// Checks if it's the slot machine emoji or the bot
	if r.Emoji.APIName() != "🎰" {
		return
	}
	if r.UserID == s.State.SessionID {
		return
	}

	// Checks if that message has a raffle react set for it and removes it
	for _, raffle := range guildRaffles {
		if raffle == nil {
			continue
		}

		if raffle.GetReactMessageID() == r.MessageID {
			db.SetGuildRaffleParticipant(r.GuildID, r.UserID, raffle, true)
			return
		}
	}
}

// Creates a raffle if it doesn't exist
func craffleCommand(s *discordgo.Session, m *discordgo.Message) {

	var temp entities.Raffle

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildRaffles := db.GetGuildRaffles(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 3)

	if len(commandStrings) != 3 ||
		(commandStrings[1] != "true" &&
			commandStrings[1] != "false") {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"craffle [react bool] [raffle name] `\n\n"+
			"Type `true` or `false` in `[react bool]` parameter to indicate whether you want users to be able to react to join the raffle. (default react emoji is slot machine.)")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	temp.SetName(strings.ToLower(commandStrings[2]))
	temp.SetParticipantIDs(nil)

	// Checks if that raffle already exists in the raffles slice
	for _, sliceRaffle := range guildRaffles {
		if sliceRaffle == nil {
			continue
		}

		if sliceRaffle.GetName() == temp.GetName() {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Such a raffle already exists.")
			if err != nil {
				common.LogError(s, guildSettings.BotLog, err)
				return
			}
			return
		}
	}

	if commandStrings[1] == "true" {
		message, err := s.ChannelMessageSend(m.ChannelID, "Raffle `"+temp.GetName()+"` is now active. ")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		err = s.MessageReactionAdd(message.ChannelID, message.ID, "🎰")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		temp.SetReactMessageID(message.ID)

		// Adds the raffle object to the raffle slice
		err = db.SetGuildRaffle(m.GuildID, &temp)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Adds the raffle object to the raffle slice
	err := db.SetGuildRaffle(m.GuildID, &temp)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Raffle `"+temp.GetName()+"` is now active. Please use `"+guildSettings.GetPrefix()+"jraffle "+temp.GetName()+"` to join the raffle.")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a raffle
func removeRaffleCommand(s *discordgo.Session, m *discordgo.Message) {
	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"removeraffle [raffle name]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	err := db.SetGuildRaffle(m.GuildID, entities.NewRaffle(commandStrings[1], nil, ""), true)
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, err.Error())
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed raffle `"+commandStrings[1]+"`.")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Picks a random winner from those participating in the raffle
func raffleWinnerCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		winnerIndex   int
		winnerID      string
		winnerMention string
	)

	guildSettings := db.GetGuildSettings(m.GuildID)

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"rafflewinner [raffle name]`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	entities.Guilds.RLock()
	entities.Guilds.DB[m.GuildID].RLock()
	for raffleIndex, raffle := range entities.Guilds.DB[m.GuildID].GetRaffles() {
		if raffle == nil {
			continue
		}

		if raffle.GetName() == strings.ToLower(commandStrings[1]) {
			participantLen := len(entities.Guilds.DB[m.GuildID].GetRaffles()[raffleIndex].GetParticipantIDs())
			if participantLen == 0 {
				winnerID = "none"
				break
			}
			winnerIndex = rand.Intn(participantLen)
			winnerID = entities.Guilds.DB[m.GuildID].GetRaffles()[raffleIndex].GetParticipantIDs()[winnerIndex]
			break
		}
	}
	entities.Guilds.RUnlock()
	entities.Guilds.DB[m.GuildID].RUnlock()

	if winnerID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such raffle exists.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	if winnerID == "none" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There is nobody to pick from to win in that raffle.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parses mention if user is in the server or not
	winnerMention = fmt.Sprintf("<@%s>", winnerID)
	_, err := s.GuildMember(m.GuildID, winnerID)
	if err != nil {
		winnerMention = common.MentionParser(s, winnerMention, m.GuildID)
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "**"+commandStrings[1]+"** winner is "+winnerMention+"! Congratulations!")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Shows existing raffles
func viewRafflesCommand(s *discordgo.Session, m *discordgo.Message) {
	var message string

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildRaffles := db.GetGuildRaffles(m.GuildID)

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.GetPrefix()+"vraffle`")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	if len(guildRaffles) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no raffles.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Iterates through all the raffles if they exist and adds them to the message string
	for _, raffle := range guildRaffles {
		if raffle == nil {
			continue
		}

		if message == "" {
			message = "Raffles:\n\n`" + raffle.GetName() + "`"
		} else {
			message += "\n `" + raffle.GetName() + "`"
		}
	}

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
	}
}

func init() {
	Add(&Command{
		Execute: raffleParticipateCommand,
		Aliases: []string{"joinraffle", "enterraffle"},
		Trigger: "jraffle",
		Desc:    "Allows you to participate in a raffle",
		Module:  "raffles",
	})
	Add(&Command{
		Execute: raffleLeaveCommand,
		Aliases: []string{"leaveraffle"},
		Trigger: "lraffle",
		Desc:    "Removes you from a raffle",
		Module:  "raffles",
	})
	Add(&Command{
		Execute:    craffleCommand,
		Aliases:    []string{"createraffle", "addraffle", "setraffle"},
		Trigger:    "craffle",
		Desc:       "Create a raffle",
		Permission: functionality.Mod,
		Module:     "raffles",
	})
	Add(&Command{
		Execute:    raffleWinnerCommand,
		Aliases:    []string{"pickrafflewin", "pickrafflewinner", "rafflewin", "winraffle", "raffwin"},
		Trigger:    "rafflewinner",
		Desc:       "Picks a random winner from those participating in a raffle",
		Permission: functionality.Mod,
		Module:     "raffles",
	})
	Add(&Command{
		Execute:    removeRaffleCommand,
		Aliases:    []string{"deleteraffle", "killraffle"},
		Trigger:    "removeraffle",
		Desc:       "Removes a previously set raffle",
		Permission: functionality.Mod,
		Module:     "raffles",
	})
	Add(&Command{
		Execute:    viewRafflesCommand,
		Aliases:    []string{"vraffles", "vraffle", "viewraffle", "raffles"},
		Trigger:    "viewraffles",
		Desc:       "Shows existing raffles",
		Permission: functionality.Mod,
		Module:     "raffles",
	})
}
