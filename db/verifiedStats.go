package db

import (
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildVerifiedStats returns the verified stats from in-memory
func GetGuildVerifiedStats(guildID string) map[string]int {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetVerifiedStats()
}

// SetGuildVerifiedStats sets a guild's verified stats in-memory
func SetGuildVerifiedStats(guildID string, verifiedStats map[string]int) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].SetVerifiedStats(verifiedStats)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("verifiedStats", entities.Guilds.DB[guildID].GetVerifiedStats())
}

// GetGuildVerifiedStat returns the amount at that date verified stats from in-memory
func GetGuildVerifiedStat(guildID string, date string) int {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	if stat, ok := entities.Guilds.DB[guildID].GetVerifiedStats()[date]; ok {
		return stat
	}

	return 0
}

// SetGuildVerifiedStat sets a guild's amount of verified stat at date in-memory
func SetGuildVerifiedStat(guildID string, date string, amount int) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].AssignToVerifiedStats(date, amount)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("verifiedStats", entities.Guilds.DB[guildID].GetVerifiedStats())
}

// AddGuildVerifiedStat adds to a guild's amount of verified stat at date in-memory
func AddGuildVerifiedStat(guildID string, date string, amount int) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].AddToVerifiedStats(date, amount)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("verifiedStats", entities.Guilds.DB[guildID].GetVerifiedStats())
}
