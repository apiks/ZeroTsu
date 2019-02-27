package commands

import (
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Assigns a user to participate in a raffle
func raffleParticipateCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		raffleExists bool
	)

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "jraffle [raffle name]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Checks if such a raffle exists and adds the user ID to it if so
	misc.MapMutex.Lock()
	for index, raffle := range misc.RafflesSlice {
		if raffle.Name == commandStrings[1] {
			raffleExists = true

			// Checks if the user already joined that raffle
			for _, ID := range raffle.ParticipantIDs {
				if ID == m.Author.ID {
					_, err := s.ChannelMessageSend(m.ChannelID, "You've already joined that raffle!")
					if err != nil {
						_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
						if err != nil {
							return
						}
						return
					}
					misc.MapMutex.Unlock()
					return
				}
			}

			// Adds user ID to the raffle list
			misc.RafflesSlice[index].ParticipantIDs = append(misc.RafflesSlice[index].ParticipantIDs, m.Author.ID)
			err := misc.RafflesWrite(misc.RafflesSlice)
			if err != nil {
				misc.CommandErrorHandler(s, m, err)
			}
			break
		}
	}
	misc.MapMutex.Unlock()
	if !raffleExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such raffle exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Success! You have entered raffle `" + commandStrings[1] + "`")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Removes a user from a raffle
func raffleLeaveCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		raffleExists bool
		userInRaffle bool
	)

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "lraffle [raffle name]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	// Checks if such a raffle exists and removes the user ID from it if so
	misc.MapMutex.Lock()
	for _, raffle := range misc.RafflesSlice {
		if raffle.Name == commandStrings[1] {
			raffleExists = true

			// Checks if the user already joined that raffle and removes him if so
			for i, ID := range raffle.ParticipantIDs {
				if ID == m.Author.ID {
					userInRaffle = true
					misc.RafflesSlice = append(misc.RafflesSlice[:i], misc.RafflesSlice[i+1:]...)
					break
				}
			}
			if !userInRaffle {
				_, err := s.ChannelMessageSend(m.ChannelID, "You're not in that raffle!")
				if err != nil {
					_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
					if err != nil {
						return
					}
					return
				}
				misc.MapMutex.Unlock()
				return
			}
			break
		}
	}
	misc.MapMutex.Unlock()
	if !raffleExists {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No such raffle exists.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "Success! You have left raffle `" + commandStrings[1] + "`")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Creates a raffle if it doesn't exist
func craffleCommand(s *discordgo.Session, m *discordgo.Message) {
	var temp misc.Raffle

	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "craffle [raffle name]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	temp.Name = commandStrings[1]
	temp.ParticipantIDs = nil

	// Checks if that raffle already exists in the raffles slice
	misc.MapMutex.Lock()
	for _, sliceRaffle := range misc.RafflesSlice {
		if sliceRaffle.Name == temp.Name {
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: Such a raffle already exists.")
			if err != nil {
				_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
				if err != nil {
					misc.MapMutex.Unlock()
					return
				}
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
	}

	// Adds the raffle object to the raffle slice
	misc.RafflesSlice = append(misc.RafflesSlice, temp)

	err := misc.RafflesWrite(misc.RafflesSlice)
	if err != nil {
		misc.MapMutex.Unlock()
		misc.CommandErrorHandler(s, m, err)
		return
	}
	misc.MapMutex.Unlock()

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Created raffle with name `" + temp.Name + "`")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Picks a random winner from those participating in the raffle
func raffleWinnerCommand(s *discordgo.Session, m *discordgo.Message) {
	var (
		winnerIndex int
		winnerID	string
	)
	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "rafflewinner [raffle name]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	misc.MapMutex.Lock()
	for raffleIndex, raffle := range misc.RafflesSlice {
		if raffle.Name == commandStrings[1] {
			participantLen := len(misc.RafflesSlice[raffleIndex].ParticipantIDs)
			winnerIndex = rand.Intn(participantLen)
			winnerID = misc.RafflesSlice[raffleIndex].ParticipantIDs[winnerIndex]
			break
		}
	}
	misc.MapMutex.Unlock()

	if winnerID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There is nobody to pick from to win in that raffle.")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	_, err := s.ChannelMessageSend(m.ChannelID, "**" + commandStrings[1] + "** winner is <@" + winnerID + ">! Congratulations!")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Removes a raffle
func removeRaffleCommand(s *discordgo.Session, m *discordgo.Message) {
	commandStrings := strings.SplitN(m.Content, " ", 2)

	if len(commandStrings) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "removeraffle [raffle name]`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	err := misc.RaffleRemove(commandStrings[1])
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, err.Error())
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Success! Removed raffle `" + commandStrings[1] + "`.")
	if err != nil {
		_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}
}

// Shows existing raffles
func viewRafflesCommand(s *discordgo.Session, m *discordgo.Message) {
	var message string

	commandStrings := strings.Split(m.Content, " ")

	if len(commandStrings) != 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `" + config.BotPrefix + "vraffle`")
		if err != nil {
			_, err = s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
		return
	}

	misc.MapMutex.Lock()
	if len(misc.RafflesSlice) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: There are no raffles.")
		if err != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, err.Error() + "\n" + misc.ErrorLocation(err))
			if err != nil {
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	// Iterates through all the raffles if they exist and adds them to the message string
	for _, raffle := range misc.RafflesSlice {
		if message == "" {
			message = "`" + raffle.Name + "`"
		} else {
			message += "\n `" + raffle.Name + "`"
		}
	}
	misc.MapMutex.Unlock()

	_, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		_, err := s.ChannelMessageSend(config.BotLogID, err.Error())
		if err != nil {
			return
		}
		return
	}
}

func init() {
	add(&command{
		execute: raffleParticipateCommand,
		aliases: []string{"joinraffle", "enterraffle"},
		trigger: "jraffle",
		desc:    "Allows you to participate in a raffle.",
		category:"misc",
	})
	add(&command{
		execute: raffleLeaveCommand,
		aliases: []string{"leaveraffle"},
		trigger: "lraffle",
		desc:    "Removes you from a raffle.",
		category:"misc",
	})
	add(&command{
		execute: craffleCommand,
		aliases: []string{"createraffle"},
		trigger: "craffle",
		desc:    "Create a raffle.",
		elevated: true,
		category:"misc",
	})
	add(&command{
		execute: raffleWinnerCommand,
		aliases: []string{"pickrafflewin", "pickrafflewinner", "rafflewin", "winraffle", "raffwin"},
		trigger: "rafflewinner",
		desc:    "Picks a random winner from those participating in a raffle.",
		elevated: true,
		category:"misc",
	})
	add(&command{
		execute: removeRaffleCommand,
		aliases: []string{"deleteraffle"},
		trigger: "removeraffle",
		desc:    "Removes a previously set raffle.",
		elevated: true,
		category:"misc",
	})
	add(&command{
		execute: viewRafflesCommand,
		aliases: []string{"vraffles", "vraffle", "viewraffle"},
		trigger: "viewraffles",
		desc:    "Shows existing raffles.",
		elevated: true,
		category:"misc",
	})
}