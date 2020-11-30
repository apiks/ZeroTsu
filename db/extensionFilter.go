package db

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/entities"
	"strings"
)

// GetGuildExtensions returns the guild extensions filters from in-memory
func GetGuildExtensions(guildID string) map[string]string {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetExtensionList()
}

// SetGuildExtension sets a target guild's extensions filters in-memory
func SetGuildExtension(guildID, extension string, deleteSlice ...bool) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	if len(deleteSlice) == 0 {
		if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetExtensionList()) >= 200 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the file extension filter limit (200) for this premium server.")
		} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetExtensionList()) >= 50 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the file extension filter (50) for this server. Please remove some or increase them to 200 by upgrading to a premium server at <https://patreon.com/animeschedule>")
		}
	}

	extension = strings.ToLower(extension)

	if len(deleteSlice) == 0 {
		var exists bool
		for guildExtension := range entities.Guilds.DB[guildID].GetExtensionList() {
			if guildExtension == "" {
				continue
			}

			if strings.ToLower(guildExtension) == extension {
				exists = true
				break
			}
		}

		if !exists {
			if entities.Guilds.DB[guildID].GetGuildSettings().GetWhitelistFileFilter() {
				entities.Guilds.DB[guildID].GetExtensionList()[strings.ToLower(extension)] = "whitelist"
			} else {
				entities.Guilds.DB[guildID].GetExtensionList()[strings.ToLower(extension)] = "blacklist"
			}
		} else {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: That extension filter already exists.")
		}
	} else {
		if _, ok := entities.Guilds.DB[guildID].GetExtensionList()[extension]; ok {
			delete(entities.Guilds.DB[guildID].GetExtensionList(), extension)
		} else {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: No such extension filter exists.")
		}
	}

	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("extensionList", entities.Guilds.DB[guildID].GetExtensionList())

	return nil
}
