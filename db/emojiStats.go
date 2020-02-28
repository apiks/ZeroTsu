package db

import (
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildEmojiStats returns the emoji stats from in-memory
func GetGuildEmojiStats(guildID string) map[string]entities.Emoji {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetEmojiStats()
}

// SetGuildEmojiStats sets a target guild's emoji stats in-memory
func SetGuildEmojiStats(guildID string, emojiStats map[string]entities.Emoji) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].SetEmojiStats(emojiStats)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("emojiStats", entities.Guilds.DB[guildID].GetEmojiStats())
}

// GetGuildEmojiStat returns a guild's emoji stat from the in-memory
func GetGuildEmojiStat(guildID, emojiID string) entities.Emoji {

	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	if stat, ok := entities.Guilds.DB[guildID].GetEmojiStats()[emojiID]; !ok {
		return stat
	}

	return entities.Emoji{}
}

// SetGuildEmojiStat sets a guild's emoji stat in-memory
func SetGuildEmojiStat(guildID string, emoji entities.Emoji) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].AssignToEmojiStats(emoji.GetID(), emoji)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("emojiStats", entities.Guilds.DB[guildID].GetEmojiStats())
}
