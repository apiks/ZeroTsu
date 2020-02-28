package db

import (
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildUserChangeStats returns the user change stats from in-memory
func GetGuildUserChangeStats(guildID string) map[string]int {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetUserChangeStats()
}

// SetGuildUserChangeStats sets a guild's user change stats in-memory
func SetGuildUserChangeStats(guildID string, userChangeStats map[string]int) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].SetUserChangeStats(userChangeStats)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("userChangeStats", entities.Guilds.DB[guildID].GetUserChangeStats())
}

// GetGuildUserChangeStat returns a guild's user change stat from the in-memory
func GetGuildUserChangeStat(guildID, date string) int {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	if stat, ok := entities.Guilds.DB[guildID].GetUserChangeStats()[date]; ok {
		return stat
	}

	return 0
}

// SetGuildUserChangeStat sets a guild's user change stat in-memory
func SetGuildUserChangeStat(guildID, date string, amount int) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].AssignToUserChangeStats(date, amount)
	entities.Guilds.DB[guildID].GetUserChangeStats()[date] += amount
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("userChangeStats", entities.Guilds.DB[guildID].GetUserChangeStats())
}

// AddGuildUserChangeStat adds to a guild's user change stat in-memory
func AddGuildUserChangeStat(guildID, date string, amount int) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].AddToUserChangeStats(date, amount)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("userChangeStats", entities.Guilds.DB[guildID].GetUserChangeStats())
}
