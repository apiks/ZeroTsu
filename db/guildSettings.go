package db

import (
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildSettings returns the guild settings from in-memory
func GetGuildSettings(guildID string) *entities.GuildSettings {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetGuildSettings()
}

// SetGuildSettings sets the guild settings from in-memory
func SetGuildSettings(guildID string, guildSettings *entities.GuildSettings) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].SetGuildSettings(guildSettings)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("guildSettings", entities.Guilds.DB[guildID].GetGuildSettings())
}
