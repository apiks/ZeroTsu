package db

import "github.com/r-anime/ZeroTsu/entities"

// GetGuildAutopost returns an autopost obect from in-memory
func GetGuildAutopost(guildID string, postType string) *entities.Cha {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	if autopost, ok := entities.Guilds.DB[guildID].GetAutoposts()[postType]; ok {
		return autopost
	}

	return nil
}

// SetGuildAutopost sets a target guild's autopost object in-memory
func SetGuildAutopost(guildID string, postType string, autopost *entities.Cha) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].GetAutoposts()[postType] = autopost
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("autoposts", entities.Guilds.DB[guildID].GetAutoposts())
}
