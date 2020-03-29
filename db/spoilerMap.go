package db

import (
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildSpoilerMap returns a guild's spoiler map from in-memory
func GetGuildSpoilerMap(guildID string) map[string]*discordgo.Role {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetSpoilerMap()
}

// SetGuildSpoilerMap sets a guild's spoiler map in-memory
func SetGuildSpoilerMap(guildID string, spoilerMap map[string]*discordgo.Role) {
	var spoilerRoles []*discordgo.Role

	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].SetSpoilerMap(spoilerMap)
	entities.Guilds.Unlock()

	for _, role := range spoilerMap {
		if role == nil {
			continue
		}
		spoilerRoles = append(spoilerRoles, role)
	}

	entities.Guilds.DB[guildID].WriteData("spoilerRoles", spoilerRoles)
}

// SetGuildSpoilerRole sets a guild's spoiler map role in-memory
func SetGuildSpoilerRole(guildID string, role *discordgo.Role, deleteSlice ...bool) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	spoilerMap := entities.Guilds.DB[guildID].GetSpoilerMap()
	if len(deleteSlice) == 0 {
		var exists bool
		for _, guildSpoilerRole := range spoilerMap {
			if guildSpoilerRole == nil {
				continue
			}

			if guildSpoilerRole.ID == role.ID {
				*guildSpoilerRole = *role
				exists = true
				break
			}
		}

		if !exists {
			spoilerMap[role.ID] = role
		}
	} else {
		delete(spoilerMap, role.ID)
	}

	entities.Guilds.Unlock()

	SetGuildSpoilerMap(guildID, spoilerMap)
}
