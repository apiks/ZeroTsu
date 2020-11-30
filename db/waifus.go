package db

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/entities"
	"strings"
)

// GetGuildWaifus the guild's waifus from in-memory
func GetGuildWaifus(guildID string) []*entities.Waifu {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetWaifus()
}

// SetGuildWaifus sets a target guild's waifus in-memory
func SetGuildWaifus(guildID string, waifus []*entities.Waifu) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(waifus) >= 400 {
		entities.Guilds.Unlock()
		return fmt.Errorf("Error: You have reached the waifu limit (400) for this premium server.")
	} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(waifus) >= 50 {
		entities.Guilds.Unlock()
		return fmt.Errorf("Error: You have reached the waifu limit (50) for this server. Please remove some or increase them to 400 by upgrading to a premium server at <https://patreon.com/animeschedule>")
	}

	entities.Guilds.DB[guildID].SetWaifus(waifus)

	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("waifus", entities.Guilds.DB[guildID].GetWaifus())

	return nil
}

// SetGuildWaifu sets a target guild's waifu in-memory
func SetGuildWaifu(guildID string, waifu entities.Waifu, delete ...bool) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	if len(delete) == 0 {
		if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetWaifus()) >= 400 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the waifu limit (400) for this premium server.")
		} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetWaifus()) >= 50 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the waifu limit (50) for this server. Please remove some or increase them to 400 by upgrading to a premium server at <https://patreon.com/animeschedule>")
		}
	}

	if len(delete) == 0 {
		var exists bool
		for _, guildWaifu := range entities.Guilds.DB[guildID].GetWaifus() {
			if guildWaifu == nil {
				continue
			}

			if strings.ToLower(guildWaifu.GetName()) == strings.ToLower(waifu.GetName()) {
				exists = true
				break
			}
		}

		if !exists {
			entities.Guilds.DB[guildID].AppendToWaifus(waifu)
		} else {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: That waifu already exists.")
		}
	} else {
		err := deleteGuildWaifu(guildID, waifu)
		if err != nil {
			entities.Guilds.Unlock()
			return err
		}
	}

	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("waifus", entities.Guilds.DB[guildID].GetWaifus())

	return nil
}

// deleteGuildWaifu safely deletes a waifu from the waifus slice
func deleteGuildWaifu(guildID string, waifu entities.Waifu) error {
	var exists bool

	for i, guildWaifu := range entities.Guilds.DB[guildID].GetWaifus() {
		if guildWaifu == nil {
			continue
		}

		if strings.ToLower(guildWaifu.GetName()) == strings.ToLower(waifu.GetName()) {
			entities.Guilds.DB[guildID].RemoveFromWaifus(i)
			exists = true
			break
		}
	}

	if !exists {
		return fmt.Errorf("Error: No such waifu exists.")
	}

	return nil
}
