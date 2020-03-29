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

// SetGuildExtensions sets a guild's extension filters in-memory
func SetGuildExtensions(guildID string, extensionList map[string]string) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].SetExtensionList(extensionList)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("extensionList", entities.Guilds.DB[guildID].GetExtensionList())
}

// SetGuildExtension sets a target guild's extensions filters in-memory
func SetGuildExtension(guildID, extension string, deleteSlice ...bool) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetExtensionList()) > 199 {
		entities.Guilds.Unlock()
		return fmt.Errorf("Error: You have reached the file extension filter limit (200) for this premium server.")
	} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetExtensionList()) > 49 {
		entities.Guilds.Unlock()
		return fmt.Errorf("Error: You have reached the file extension filter (50) for this server. Please remove some or increase them to 200 by upgrading to a premium server at <https://patreon.com/apiks>")
	}

	extension = strings.ToLower(extension)
	extensionList := entities.Guilds.DB[guildID].GetExtensionList()

	if len(deleteSlice) == 0 {
		var exists bool
		for guildExtension := range extensionList {
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
				extensionList[strings.ToLower(extension)] = "whitelist"
			} else {
				extensionList[strings.ToLower(extension)] = "blacklist"
			}
		} else {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: That extension filter already exists.")
		}
	} else {
		if _, ok := extensionList[extension]; ok {
			delete(extensionList, extension)
		} else {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: No such extension filter exists.")
		}
	}
	entities.Guilds.Unlock()

	SetGuildExtensions(guildID, extensionList)
	return nil
}
