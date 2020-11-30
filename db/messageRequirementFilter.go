package db

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/entities"
	"strings"
)

// GetGuildMessageRequirements returns the guild's message requirement filters
func GetGuildMessageRequirements(guildID string) []entities.MessRequirement {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetMessageRequirements()
}

// SetGuildMessageRequirements sets a target guild's message requirement filters in-memory
func SetGuildMessageRequirements(guildID string, messRequirements []entities.MessRequirement) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(messRequirements) >= 150 {
		entities.Guilds.Unlock()
		return fmt.Errorf("Error: You have reached the message requirement filter limit (150) for this premium server.")
	} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(messRequirements) >= 50 {
		entities.Guilds.Unlock()
		return fmt.Errorf("Error: You have reached the message requirement filter limit (50) for this server. Please remove some or increase them to 150 by upgrading to a premium server at <https://patreon.com/animeschedule>")
	}

	entities.Guilds.DB[guildID].SetMessageRequirements(messRequirements)

	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("messReqs", entities.Guilds.DB[guildID].GetMessageRequirements())

	return nil
}

// SetGuildMessageRequirement sets a target guild's message requirement filter in-memory
func SetGuildMessageRequirement(guildID string, messRequirement entities.MessRequirement, delete ...bool) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	if len(delete) == 0 {
		if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetMessageRequirements()) >= 150 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the message requirement filter limit (150) for this premium server.")
		} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetMessageRequirements()) >= 50 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the message requirement filter limit (50) for this server. Please remove some or increase them to 150 by upgrading to a premium server at <https://patreon.com/animeschedule>")
		}
	}

	messRequirement = messRequirement.SetPhrase(strings.ToLower(messRequirement.GetPhrase()))

	if len(delete) == 0 {
		var exists bool
		for _, guildMessReq := range entities.Guilds.DB[guildID].GetMessageRequirements() {
			if strings.ToLower(guildMessReq.GetPhrase()) == messRequirement.GetPhrase() {
				exists = true
				break
			}
		}

		if !exists {
			entities.Guilds.DB[guildID].AppendToMessageRequirements(messRequirement)
		} else {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: That message requirement already exists.")
		}
	} else {
		err := deleteGuildMessageRequirement(guildID, messRequirement)
		if err != nil {
			entities.Guilds.Unlock()
			return err
		}
	}

	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("messReqs", entities.Guilds.DB[guildID].GetMessageRequirements())

	return nil
}

// deleteGuildMessageRequirement safely deletes a message requirement from the message requirements slice
func deleteGuildMessageRequirement(guildID string, messReq entities.MessRequirement) error {
	var exists bool

	for i, guildMessReq := range entities.Guilds.DB[guildID].GetMessageRequirements() {
		if strings.ToLower(guildMessReq.GetPhrase()) == messReq.GetPhrase() {
			entities.Guilds.DB[guildID].RemoveFromMessageRequirements(i)
			exists = true
			break
		}
	}

	if !exists {
		return fmt.Errorf("Error: No such message requirement exists.")
	}

	return nil
}
