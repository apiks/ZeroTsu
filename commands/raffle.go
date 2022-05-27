package commands

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// raffleParticipateCommand assigns a user to participate in a raffle
func raffleParticipateCommand(raffleName, authorID, guildID string) string {
	var (
		raffleExists bool
		guildRaffles = db.GetGuildRaffles(guildID)
	)

	// Checks if such a raffle exists and adds the user ID to it if so
	for _, raffle := range guildRaffles {
		if raffle == nil {
			continue
		}

		if strings.ToLower(raffle.GetName()) == strings.ToLower(raffleName) {
			raffleExists = true

			// Checks if the user already joined that raffle
			for _, ID := range raffle.GetParticipantIDs() {
				if ID != authorID {
					continue
				}
				return "Error: You've already joined that raffle!"
			}

			// Adds user ID to the raffle list
			db.SetGuildRaffleParticipant(guildID, authorID, raffle)
			break
		}
	}

	if !raffleExists {
		return "Error: No such raffle exists."
	}

	return fmt.Sprintf("Success! You have entered raffle `%s`", raffleName)
}

// raffleParticipateCommandHandler assigns a user to participate in a raffle
func raffleParticipateCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	var raffleExists bool

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

// raffleLeaveCommand removes a user from a raffle
func raffleLeaveCommand(raffleName, authorID, guildID string) string {
	var (
		raffleExists bool
		userInRaffle bool
		guildRaffles = db.GetGuildRaffles(guildID)
	)

	// Checks if such a raffle exists and removes the user ID from it if so
	for _, raffle := range guildRaffles {
		if raffle == nil {
			continue
		}

		if strings.ToLower(raffle.GetName()) == strings.ToLower(raffleName) {
			raffleExists = true

			// Checks if the user already joined that raffle and removes him if so
			for _, ID := range raffle.GetParticipantIDs() {
				if ID == authorID {
					userInRaffle = true

					// Leaves the raffle
					db.SetGuildRaffleParticipant(guildID, authorID, raffle, true)
					break
				}
			}
			if !userInRaffle {
				return "Error: You haven't joined that raffle!"
			}
			break
		}
	}

	if !raffleExists {
		return "Error: No such raffle exists."
	}

	return fmt.Sprintf("Success! You have left raffle `%s`", raffleName)
}

// raffleLeaveCommandHandler removes a user from a raffle
func raffleLeaveCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

// RaffleReactJoinHandler enters a user in a raffle if they react
func RaffleReactJoinHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in RaffleReactJoinHandler")
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

// RaffleReactLeaveHandler removes a user from a raffle if they unreact
func RaffleReactLeaveHandler(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in RaffleReactLeaveHandler")
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

// craffleCommand creates a raffle if it doesn't exist
func craffleCommand(s *discordgo.Session, raffleName string, reactJoin bool, channelID, guildID string) string {
	var (
		temp         entities.Raffle
		guildRaffles = db.GetGuildRaffles(guildID)
	)

	temp.SetName(strings.ToLower(raffleName))
	temp.SetParticipantIDs(nil)

	// Checks if that raffle already exists in the raffles slice
	for _, sliceRaffle := range guildRaffles {
		if sliceRaffle == nil {
			continue
		}

		if sliceRaffle.GetName() == temp.GetName() {
			return "Error: A raffle with this name already exists."
		}
	}

	if reactJoin {
		message, err := s.ChannelMessageSend(channelID, fmt.Sprintf("Raffle `%s` is now active.", raffleName))
		if err != nil {
			return err.Error()
		}
		err = s.MessageReactionAdd(message.ChannelID, message.ID, "🎰")
		if err != nil {
			return err.Error()
		}
		temp.SetReactMessageID(message.ID)

		// Adds the raffle object to the raffle slice
		err = db.SetGuildRaffle(guildID, &temp)
		if err != nil {
			return err.Error()
		}

		return ""
	}

	// Adds the raffle object to the raffle slice
	err := db.SetGuildRaffle(guildID, &temp)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Raffle `%s` is now active.", raffleName)
}

// craffleCommandHandler creates a raffle if it doesn't exist
func craffleCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

// removeRaffleCommand removes a raffle
func removeRaffleCommand(raffleName, guildID string) string {
	err := db.SetGuildRaffle(guildID, entities.NewRaffle(raffleName, nil, ""), true)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Success! Removed raffle `%s`", raffleName)
}

// removeRaffleCommandHandler removes a raffle
func removeRaffleCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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

// raffleWinnerCommand picks a random winner from those participating in the raffle
func raffleWinnerCommand(raffleName, guildID string) string {
	var (
		winnerIndex int
		winnerID    string
	)

	entities.Guilds.RLock()
	entities.Guilds.DB[guildID].RLock()
	for raffleIndex, raffle := range entities.Guilds.DB[guildID].GetRaffles() {
		if raffle == nil {
			continue
		}

		if strings.ToLower(raffle.GetName()) == strings.ToLower(raffleName) {
			participantLen := len(entities.Guilds.DB[guildID].GetRaffles()[raffleIndex].GetParticipantIDs())
			if participantLen == 0 {
				winnerID = "none"
				break
			}
			winnerIndex = rand.Intn(participantLen)
			winnerID = entities.Guilds.DB[guildID].GetRaffles()[raffleIndex].GetParticipantIDs()[winnerIndex]
			break
		}
	}
	entities.Guilds.RUnlock()
	entities.Guilds.DB[guildID].RUnlock()

	if winnerID == "" {
		return "Error: No such raffle exists."
	}
	if winnerID == "none" {
		return "Error: There is nobody to pick to win in that raffle."
	}

	return fmt.Sprintf("**%s** winner is %s! Congratulations!", raffleName, fmt.Sprintf("<@%s>", winnerID))
}

// raffleWinnerCommandHandler picks a random winner from those participating in the raffle
func raffleWinnerCommandHandler(s *discordgo.Session, m *discordgo.Message) {
	var (
		winnerIndex int
		winnerID    string
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

	_, err := s.ChannelMessageSend(m.ChannelID, "**"+commandStrings[1]+"** winner is "+fmt.Sprintf("<@%s>", winnerID)+"! Congratulations!")
	if err != nil {
		common.LogError(s, guildSettings.BotLog, err)
		return
	}
}

// viewRafflesCommand prints existing raffles
func viewRafflesCommand(guildID string) []string {
	var (
		message      string
		guildRaffles = db.GetGuildRaffles(guildID)
	)

	if len(guildRaffles) == 0 {
		return []string{"Error: There are no raffles."}
	}

	// Iterates through all the raffles if they exist and adds them to the message string
	for _, raffle := range guildRaffles {
		if raffle == nil {
			continue
		}

		if message == "" {
			message = "Raffles:\n`" + raffle.GetName() + "`"
		} else {
			message += "\n `" + raffle.GetName() + "`"
		}
	}

	return common.SplitLongMessage(message)
}

// viewRafflesCommandHandler prints existing raffles
func viewRafflesCommandHandler(s *discordgo.Session, m *discordgo.Message) {
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
			message = "Raffles:\n`" + raffle.GetName() + "`"
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
		Execute: raffleParticipateCommandHandler,
		Name:    "join-raffle",
		Aliases: []string{"joinraffle", "enterraffle", "jraffle", "j-raffle"},
		Desc:    "Enters you in a raffle",
		Module:  "raffles",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "raffle",
				Description: "The name of the raffle you want to enter.",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var raffleName string
			if i.ApplicationCommandData().Options == nil {
				return
			}

			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "raffle" {
					raffleName = option.StringValue()
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: raffleParticipateCommand(raffleName, i.Member.User.ID, i.GuildID),
				},
			})
		},
	})
	Add(&Command{
		Execute: raffleLeaveCommandHandler,
		Name:    "leave-raffle",
		Aliases: []string{"leaveraffle", "lraffle", "l-raffle"},
		Desc:    "Removes you from a raffle",
		Module:  "raffles",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "raffle",
				Description: "The name of the raffle you want to leave.",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var raffleName string
			if i.ApplicationCommandData().Options == nil {
				return
			}

			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "raffle" {
					raffleName = option.StringValue()
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: raffleLeaveCommand(raffleName, i.Member.User.ID, i.GuildID),
				},
			})
		},
	})
	Add(&Command{
		Execute:    craffleCommandHandler,
		Name:       "create-raffle",
		Aliases:    []string{"createraffle", "addraffle", "setraffle", "craffle", "c-raffle"},
		Desc:       "Creates a raffle",
		Permission: functionality.Mod,
		Module:     "raffles",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "raffle",
				Description: "The name of the raffle you want to leave.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "react",
				Description: "Whether you want users to be able to use a slot emoji reaction to enter/leave the raffle.",
				Required:    false,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "create-raffle", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			var (
				raffleName  string
				raffleReact bool
			)
			if i.ApplicationCommandData().Options == nil {
				return
			}

			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "raffle" {
					raffleName = option.StringValue()
				} else if option.Name == "react" {
					raffleReact = option.BoolValue()
				}
			}

			if raffleReact {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Creating raffle. . .",
					},
				})

				_ = craffleCommand(s, raffleName, raffleReact, i.ChannelID, i.GuildID)
			} else {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: craffleCommand(nil, raffleName, false, "", i.GuildID),
					},
				})
			}
		},
	})
	Add(&Command{
		Execute:    raffleWinnerCommandHandler,
		Name:       "pick-raffle-winner",
		Aliases:    []string{"pickrafflewin", "pickrafflewinner", "rafflewin", "winraffle", "raffwin", "rafflewinner"},
		Desc:       "Picks a random winner from those participating in a given raffle",
		Permission: functionality.Mod,
		Module:     "raffles",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "raffle",
				Description: "The name of the raffle you want select a winner in.",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "pick-raffle-winner", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			var raffleName string
			if i.ApplicationCommandData().Options == nil {
				return
			}

			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "raffle" {
					raffleName = option.StringValue()
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: raffleWinnerCommand(raffleName, i.GuildID),
				},
			})
		},
	})
	Add(&Command{
		Execute:    removeRaffleCommandHandler,
		Name:       "remove-raffle",
		Aliases:    []string{"deleteraffle", "killraffle", "removeraffle", "delete-raffle", "removeraffle"},
		Desc:       "Removes a set raffle",
		Permission: functionality.Mod,
		Module:     "raffles",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "raffle",
				Description: "The name of the raffle you want to remove.",
				Required:    true,
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "remove-raffle", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			var raffleName string
			if i.ApplicationCommandData().Options == nil {
				return
			}

			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "raffle" {
					raffleName = option.StringValue()
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: removeRaffleCommand(raffleName, i.GuildID),
				},
			})
		},
	})
	Add(&Command{
		Execute:    viewRafflesCommandHandler,
		Name:       "show-raffles",
		Aliases:    []string{"vraffles", "vraffle", "viewraffle", "raffles", "view-raffles", "viewraffles", "printraffles", "print-raffles", "showraffles"},
		Desc:       "Prints existing raffles",
		Permission: functionality.Mod,
		Module:     "raffles",
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := VerifySlashCommand(s, "show-raffles", i)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: err.Error(),
					},
				})
				return
			}

			messages := viewRafflesCommand(i.GuildID)
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

					s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
						Content: message,
					})
				}
			}
		},
	})
}
