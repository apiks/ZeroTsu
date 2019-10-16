package commands

import (
	"fmt"
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

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"jraffle [raffle name]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if such a raffle exists and adds the user ID to it if so
	functionality.MapMutex.Lock()
	for index, raffle := range functionality.GuildMap[m.GuildID].Raffles {
		if raffle.Name == strings.ToLower(commandStrings[1]) {
			raffleExists = true

			// Checks if the user already joined that raffle
			for _, ID := range raffle.ParticipantIDs {
				if ID == m.Author.ID {
					_, err := s.ChannelMessageSend(m.ChannelID, "You've already joined that raffle!")
					if err != nil {
						functionality.MapMutex.Unlock()
						functionality.LogError(s, guildSettings.BotLog, err)
						return
					}
					functionality.MapMutex.Unlock()
					return
				}
			}

			// Adds user ID to the raffle list
			functionality.GuildMap[m.GuildID].Raffles[index].ParticipantIDs = append(functionality.GuildMap[m.GuildID].Raffles[index].ParticipantIDs, m.Author.ID)
			_ = functionality.RafflesWrite(functionality.GuildMap[m.GuildID].Raffles, m.GuildID)
			break
		}
	}
	functionality.MapMutex.Unlock()
	if !raffleExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such raffle exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Success! You have entered raffle `"+commandStrings[1]+"`")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a user from a raffle
func raffleLeaveCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		raffleExists bool
		userInRaffle bool
	)

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"lraffle [raffle name]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Checks if such a raffle exists and removes the user ID from it if so
	functionality.MapMutex.Lock()
	for _, raffle := range functionality.GuildMap[m.GuildID].Raffles {
		if raffle.Name == strings.ToLower(commandStrings[1]) {
			raffleExists = true

			// Checks if the user already joined that raffle and removes him if so
			for i, ID := range raffle.ParticipantIDs {
				if ID == m.Author.ID {
					userInRaffle = true
					functionality.GuildMap[m.GuildID].Raffles = append(functionality.GuildMap[m.GuildID].Raffles[:i], functionality.GuildMap[m.GuildID].Raffles[i+1:]...)
					break
				}
			}
			if !userInRaffle {
				_, err := s.ChannelMessageSend(m.ChannelID, "You're not in that raffle!")
				if err != nil {
					functionality.MapMutex.Unlock()
					functionality.LogError(s, guildSettings.BotLog, err)
					return
				}
				functionality.MapMutex.Unlock()
				return
			}
			break
		}
	}
	functionality.MapMutex.Unlock()
	if !raffleExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such raffle exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Success! You have left raffle `"+commandStrings[1]+"`")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
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

	functionality.MapMutex.Lock()
	functionality.HandleNewGuild(s, r.GuildID)

	// Checks if it's the slot machine emoji or the bot itself
	if r.Emoji.APIName() != "🎰" {
		functionality.MapMutex.Unlock()
		return
	}
	if r.UserID == s.State.User.ID {
		functionality.MapMutex.Unlock()
		return
	}

	// Checks if that message has a raffle react set for it
	for i, raffle := range functionality.GuildMap[r.GuildID].Raffles {
		if raffle.ReactMessageID == r.MessageID {
			functionality.GuildMap[r.GuildID].Raffles[i].ParticipantIDs = append(functionality.GuildMap[r.GuildID].Raffles[i].ParticipantIDs, r.UserID)
			_ = functionality.RafflesWrite(functionality.GuildMap[r.GuildID].Raffles, r.GuildID)
			functionality.MapMutex.Unlock()
			return
		}
	}
	functionality.MapMutex.Unlock()
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

	functionality.MapMutex.Lock()
	functionality.HandleNewGuild(s, r.GuildID)

	// Checks if it's the slot machine emoji or the bot
	if r.Emoji.APIName() != "🎰" {
		functionality.MapMutex.Unlock()
		return
	}
	if r.UserID == s.State.SessionID {
		functionality.MapMutex.Unlock()
		return
	}

	// Checks if that message has a raffle react set for it and removes it
	for index, raffle := range functionality.GuildMap[r.GuildID].Raffles {
		if raffle.ReactMessageID == r.MessageID {
			for i := range functionality.GuildMap[r.GuildID].Raffles[index].ParticipantIDs {
				functionality.GuildMap[r.GuildID].Raffles[index].ParticipantIDs = append(functionality.GuildMap[r.GuildID].Raffles[index].ParticipantIDs[:i], functionality.GuildMap[r.GuildID].Raffles[index].ParticipantIDs[i+1:]...)
			}
			_ = functionality.RafflesWrite(functionality.GuildMap[r.GuildID].Raffles, r.GuildID)
			functionality.MapMutex.Unlock()
			return
		}
	}
	functionality.MapMutex.Unlock()
}

// Creates a raffle if it doesn't exist
func craffleCommand(s *discordgo.Session, m *discordgo.Message) {

	var temp functionality.Raffle

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 3)

	if len(commandStrings) != 3 ||
		(commandStrings[1] != "true" &&
			commandStrings[1] != "false") {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"craffle [react bool] [raffle name] `\n\n"+
			"Type `true` or `false` in `[react bool]` parameter to indicate whether you want users to be able to react to join the raffle. (default react emoji is slot machine.)")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	temp.Name = strings.ToLower(commandStrings[2])
	temp.ParticipantIDs = nil

	// Checks if that raffle already exists in the raffles slice
	functionality.MapMutex.Lock()
	for _, sliceRaffle := range functionality.GuildMap[m.GuildID].Raffles {
		if sliceRaffle.Name == temp.Name {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Such a raffle already exists.")
			if err != nil {
				functionality.MapMutex.Unlock()
				functionality.LogError(s, guildSettings.BotLog, err)
				return
			}
			functionality.MapMutex.Unlock()
			return
		}
	}
	functionality.MapMutex.Unlock()

	if commandStrings[1] == "true" {
		message, err := s.ChannelMessageSend(m.ChannelID, "Raffle `"+temp.Name+"` is now active. ")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		err = s.MessageReactionAdd(message.ChannelID, message.ID, "🎰")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		temp.ReactMessageID = message.ID

		// Adds the raffle object to the raffle slice
		functionality.MapMutex.Lock()
		functionality.GuildMap[m.GuildID].Raffles = append(functionality.GuildMap[m.GuildID].Raffles, &temp)

		// Writes the raffle object to storage
		err = functionality.RafflesWrite(functionality.GuildMap[m.GuildID].Raffles, m.GuildID)
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}

	// Adds the raffle object to the raffle slice
	functionality.MapMutex.Lock()
	functionality.GuildMap[m.GuildID].Raffles = append(functionality.GuildMap[m.GuildID].Raffles, &temp)

	// Writes the raffle object to storage
	err := functionality.RafflesWrite(functionality.GuildMap[m.GuildID].Raffles, m.GuildID)
	if err != nil {
		functionality.MapMutex.Unlock()
		functionality.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}
	functionality.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, "Raffle `"+temp.Name+"` is now active. Please use `"+guildSettings.Prefix+"jraffle "+temp.Name+"` to join the raffle.")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Removes a raffle
func removeRaffleCommand(s *discordgo.Session, m *discordgo.Message) {

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"removeraffle [raffle name]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	err := functionality.RaffleRemove(commandStrings[1], m.GuildID)
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, err.Error())
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed raffle `"+commandStrings[1]+"`.")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
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

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.SplitN(strings.Replace(m.Content, "  ", " ", -1), " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"rafflewinner [raffle name]`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	functionality.MapMutex.Lock()
	for raffleIndex, raffle := range functionality.GuildMap[m.GuildID].Raffles {
		if raffle.Name == strings.ToLower(commandStrings[1]) {
			participantLen := len(functionality.GuildMap[m.GuildID].Raffles[raffleIndex].ParticipantIDs)
			if participantLen == 0 {
				winnerID = "none"
				break
			}
			winnerIndex = rand.Intn(participantLen)
			winnerID = functionality.GuildMap[m.GuildID].Raffles[raffleIndex].ParticipantIDs[winnerIndex]
			break
		}
	}
	functionality.MapMutex.Unlock()

	if winnerID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such raffle exists.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}
	if winnerID == "none" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There is nobody to pick from to win in that raffle.")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Parses mention if user is in the server or not
	winnerMention = fmt.Sprintf("<@%v>", winnerID)
	_, err := s.GuildMember(m.GuildID, winnerID)
	if err != nil {
		winnerMention = functionality.MentionParser(s, winnerMention, m.GuildID)
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "**"+commandStrings[1]+"** winner is "+winnerMention+"! Congratulations!")
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// Shows existing raffles
func viewRafflesCommand(s *discordgo.Session, m *discordgo.Message) {

	var message string

	functionality.MapMutex.Lock()
	guildSettings := functionality.GuildMap[m.GuildID].GetGuildSettings()
	functionality.MapMutex.Unlock()

	commandStrings := strings.Split(strings.Replace(m.Content, "  ", " ", -1), " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `"+guildSettings.Prefix+"vraffle`")
		if err != nil {
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	functionality.MapMutex.Lock()
	if len(functionality.GuildMap[m.GuildID].Raffles) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no raffles.")
		if err != nil {
			functionality.MapMutex.Unlock()
			functionality.LogError(s, guildSettings.BotLog, err)
			return
		}
		functionality.MapMutex.Unlock()
		return
	}

	// Iterates through all the raffles if they exist and adds them to the message string
	for _, raffle := range functionality.GuildMap[m.GuildID].Raffles {
		if message == "" {
			message = "Raffles:\n\n`" + raffle.Name + "`"
		} else {
			message += "\n `" + raffle.Name + "`"
		}
	}
	functionality.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		functionality.LogError(s, guildSettings.BotLog, err)
		return
	}
}

func init() {
	functionality.Add(&functionality.Command{
		Execute: raffleParticipateCommand,
		Aliases: []string{"joinraffle", "enterraffle"},
		Trigger: "jraffle",
		Desc:    "Allows you to participate in a raffle",
		Module:  "raffles",
	})
	functionality.Add(&functionality.Command{
		Execute: raffleLeaveCommand,
		Aliases: []string{"leaveraffle"},
		Trigger: "lraffle",
		Desc:    "Removes you from a raffle",
		Module:  "raffles",
	})
	functionality.Add(&functionality.Command{
		Execute:    craffleCommand,
		Aliases:    []string{"createraffle", "addraffle", "setraffle"},
		Trigger:    "craffle",
		Desc:       "Create a raffle",
		Permission: functionality.Mod,
		Module:     "raffles",
	})
	functionality.Add(&functionality.Command{
		Execute:    raffleWinnerCommand,
		Aliases:    []string{"pickrafflewin", "pickrafflewinner", "rafflewin", "winraffle", "raffwin"},
		Trigger:    "rafflewinner",
		Desc:       "Picks a random winner from those participating in a raffle",
		Permission: functionality.Mod,
		Module:     "raffles",
	})
	functionality.Add(&functionality.Command{
		Execute:    removeRaffleCommand,
		Aliases:    []string{"deleteraffle", "killraffle"},
		Trigger:    "removeraffle",
		Desc:       "Removes a previously set raffle",
		Permission: functionality.Mod,
		Module:     "raffles",
	})
	functionality.Add(&functionality.Command{
		Execute:    viewRafflesCommand,
		Aliases:    []string{"vraffles", "vraffle", "viewraffle", "raffles"},
		Trigger:    "viewraffles",
		Desc:       "Shows existing raffles",
		Permission: functionality.Mod,
		Module:     "raffles",
	})
}
