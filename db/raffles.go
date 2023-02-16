package db

import (
	"fmt"
	"strings"

	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildRaffles the guild's raffles from in-memory
func GetGuildRaffles(guildID string) []*entities.Raffle {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetRaffles()
}

// SetGuildRaffle sets a target guild's raffle in-memory
func SetGuildRaffle(guildID string, raffle *entities.Raffle, delete ...bool) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	defer entities.Guilds.Unlock()

	if len(delete) == 0 {
		if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetRaffles()) >= 200 {
			return fmt.Errorf("Error: You have reached the raffle limit (200) for this premium server.")
		} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetRaffles()) >= 50 {
			return fmt.Errorf("Error: You have reached the raffle limit (50) for this server. Please remove some or increase them to 200 by upgrading to a premium server at <https://patreon.com/animeschedule>")
		}
	}

	raffle.SetName(strings.ToLower(raffle.GetName()))

	if len(delete) == 0 {
		var exists bool
		for _, guildRaffle := range entities.Guilds.DB[guildID].GetRaffles() {
			if strings.ToLower(guildRaffle.GetName()) == raffle.GetName() {
				exists = true
				break
			}
		}

		if !exists {
			entities.Guilds.DB[guildID].AppendToRaffles(raffle)
		} else {
			return fmt.Errorf("Error: That raffle already exists.")
		}
	} else {
		err := deleteGuildRaffle(guildID, raffle)
		if err != nil {
			return err
		}
	}

	entities.Guilds.DB[guildID].WriteData("raffles", entities.Guilds.DB[guildID].GetRaffles())

	return nil
}

// deleteGuildRaffle safely deletes a raffle from the raffles slice
func deleteGuildRaffle(guildID string, raffle *entities.Raffle) error {
	var exists bool

	for i, guildRaffle := range entities.Guilds.DB[guildID].GetRaffles() {
		if guildRaffle == nil {
			continue
		}

		if strings.ToLower(guildRaffle.GetName()) == raffle.GetName() {
			entities.Guilds.DB[guildID].RemoveFromRaffles(i)
			exists = true
			break
		}
	}

	if !exists {
		return fmt.Errorf("Error: No such raffle exists.")
	}

	return nil
}

// SetGuildRaffleParticipant sets a target guild's raffle participant in-memory
func SetGuildRaffleParticipant(guildID, userID string, raffle *entities.Raffle, delete ...bool) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	defer entities.Guilds.Unlock()

	raffle.SetName(strings.ToLower(raffle.GetName()))

	if len(delete) == 0 {
		raffle.AppendToParticipantIDs(userID)
		for _, guildRaffle := range entities.Guilds.DB[guildID].GetRaffles() {
			if guildRaffle == nil {
				continue
			}

			if strings.ToLower(guildRaffle.GetName()) == raffle.GetName() {
				*guildRaffle = *raffle
				break
			}
		}
	} else {
		deleteGuildRaffleParticipant(userID, raffle)
	}

	entities.Guilds.DB[guildID].WriteData("raffles", entities.Guilds.DB[guildID].GetRaffles())

	return
}

// deleteGuildRaffleParticipant safely deletes a raffle participant from the raffles participantIds slice
func deleteGuildRaffleParticipant(userID string, raffle *entities.Raffle) {
	for i, participantID := range raffle.GetParticipantIDs() {
		if participantID == userID {
			raffle.RemoveFromParticipantIDs(i)
			break
		}
	}
}
