package db

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildWaifuTrades the guild's waifu trades from in-memory
func GetGuildWaifuTrades(guildID string) []*entities.WaifuTrade {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetWaifuTrades()
}

// SetGuildWaifuTrades sets a target guild's waifu trades in-memory
func SetGuildWaifuTrades(guildID string, waifuTrades []*entities.WaifuTrade) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(waifuTrades) >= 800 {
		entities.Guilds.Unlock()
		return fmt.Errorf("Error: This premium server has reached the waifu trade limit (800).")
	} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(waifuTrades) >= 400 {
		entities.Guilds.Unlock()
		return fmt.Errorf("Error: This server has reached the waifu trade limit (400). Please contact the bot creator or increase the limit to 500 by upgrading to a premium server at <https://patreon.com/animeschedule>")
	}

	entities.Guilds.DB[guildID].SetWaifuTrades(waifuTrades)

	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("waifuTrades", entities.Guilds.DB[guildID].GetWaifuTrades())

	return nil
}

// SetGuildWaifuTrade sets a guild waifu trade in-memory
func SetGuildWaifuTrade(guildID string, waifuTrade *entities.WaifuTrade, delete ...bool) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	if len(delete) == 0 {
		if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetWaifuTrades()) >= 800 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: This premium server has reached the waifu trade limit (800).")
		} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetWaifuTrades()) >= 400 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: This server has reached the waifu trade limit (400). Please contact the bot creator or increase the limit to 500 by upgrading to a premium server at <https://patreon.com/animeschedule>")
		}
	}

	if len(delete) == 0 {
		var exists bool
		for _, guildWaifuTrade := range entities.Guilds.DB[guildID].GetWaifuTrades() {
			if guildWaifuTrade == nil {
				continue
			}

			if *guildWaifuTrade == *waifuTrade {
				exists = true
			}
		}

		if !exists {
			entities.Guilds.DB[guildID].AppendToWaifuTrades(waifuTrade)
		} else {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: That waifu trade already exists.")
		}
	} else {
		err := deleteGuildWaifuTrade(guildID, waifuTrade)
		if err != nil {
			entities.Guilds.Unlock()
			return err
		}
	}

	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("waifuTrades", entities.Guilds.DB[guildID].GetWaifuTrades())

	return nil
}

// deleteGuildWaifuTrade safely deletes a waifu trade from the waifuTrades slice
func deleteGuildWaifuTrade(guildID string, waifuTrade *entities.WaifuTrade) error {
	var exists bool

	for i, guildWaifuTrade := range entities.Guilds.DB[guildID].GetWaifuTrades() {
		if guildWaifuTrade == nil {
			continue
		}

		if *guildWaifuTrade == *waifuTrade {
			entities.Guilds.DB[guildID].RemoveFromWaifuTrades(i)
			exists = true
			break
		}
	}

	if !exists {
		return fmt.Errorf("Error: No such waifu trade exists.")
	}

	return nil
}
