package db

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildTempChannels returns a guild's temporary channels from in-memory
func GetGuildTempChannels(guildID string) map[string]*entities.TempChaInfo {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetTempChaMap()
}

// SetGuildTempChannels sets a guild's temporary channels in-memory
func SetGuildTempChannels(guildID string, tempCha map[string]*entities.TempChaInfo) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].SetTempChaMap(tempCha)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("tempCha", entities.Guilds.DB[guildID].GetTempChaMap())
}

// SetGuildTempChannel sets a guild's temporary channel in-memory
func SetGuildTempChannel(guildID, roleID string, tempCha *entities.TempChaInfo, deleteSlice ...bool) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	if len(deleteSlice) == 0 {
		if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetTempChaMap()) >= 200 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the temporary channel limit (200) for this premium server.")
		} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetTempChaMap()) >= 50 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the temporary channel limit (50) for this server. Please wait for some to be removed or increase them to 200 by upgrading to a premium server at <https://patreon.com/animeschedule>")
		}
	}

	if len(deleteSlice) == 0 {
		var exists bool
		for _, guildTempCha := range entities.Guilds.DB[guildID].GetTempChaMap() {
			if guildTempCha == nil {
				continue
			}

			if guildTempCha.RoleName == tempCha.RoleName &&
				guildTempCha.CreationDate == tempCha.CreationDate &&
				guildTempCha.Elevated == tempCha.Elevated {
				*guildTempCha = *tempCha
				exists = true
				break
			}
		}

		if !exists {
			entities.Guilds.DB[guildID].GetTempChaMap()[roleID] = tempCha
		}
	} else {
		if _, ok := entities.Guilds.DB[guildID].GetTempChaMap()[roleID]; ok {
			delete(entities.Guilds.DB[guildID].GetTempChaMap(), roleID)
		} else {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: That temporary channel doesn't exist.")
		}
	}

	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("tempCha", entities.Guilds.DB[guildID].GetTempChaMap())

	return nil
}
