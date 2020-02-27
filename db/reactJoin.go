package db

import (
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildReactJoin returns a guild's react join map from in-memory
func GetGuildReactJoin(guildID string) map[string]*entities.ReactJoin {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return  entities.Guilds.DB[guildID].GetReactJoinMap()
}

// SetGuildReactJoin sets a guild's react join map in-memory
func SetGuildReactJoin(guildID string, reactJoin map[string]*entities.ReactJoin) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].SetReactJoinMap(reactJoin)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("reactJoin", reactJoin)
}

// SetGuildReactJoinEmoji sets a guild's react join emoji map in-memory
func SetGuildReactJoinEmoji(guildID, messageID string, reactJoinEmoji *entities.ReactJoin) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].ReactJoinMap[messageID] = reactJoinEmoji
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("reactJoin", entities.Guilds.DB[guildID].GetReactJoinMap())
}
