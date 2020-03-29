package db

import "github.com/r-anime/ZeroTsu/entities"

// GetGuildAutopost returns an autopost obect from in-memory
func GetGuildAutopost(guildID string, postType string) entities.Cha {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	if autopost, ok := entities.Guilds.DB[guildID].GetAutoposts()[postType]; ok {
		return autopost
	}

	return entities.Cha{}
}

// SetGuildAutoposts sets a guild's autoposts in-memory
func SetGuildAutoposts(guildID string, autoposts map[string]entities.Cha) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].SetAutoposts(autoposts)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("autoposts", entities.Guilds.DB[guildID].GetAutoposts())
}

// SetGuildAutopost sets a target guild's autopost object in-memory
func SetGuildAutopost(guildID string, postType string, autopost entities.Cha) {
	entities.HandleNewGuild(guildID)

	autoposts := entities.Guilds.DB[guildID].GetAutoposts()
	entities.Guilds.Lock()
	autoposts[postType] = autopost
	entities.Guilds.Unlock()

	SetGuildAutoposts(guildID, autoposts)
}
