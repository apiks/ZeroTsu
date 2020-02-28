package db

import (
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildChannelStats returns the channel stats from in-memory
func GetGuildChannelStats(guildID string) map[string]entities.Channel {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetChannelStats()
}

// SetGuildChannelStats sets a target guild's channel stats in-memory
func SetGuildChannelStats(guildID string, channelStats map[string]entities.Channel) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].SetChannelStats(channelStats)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("channelStats", entities.Guilds.DB[guildID].GetChannelStats())
}

// GetGuildEmojiStat returns a guild's emoji stat from the in-memory
func GetGuildChannelStat(guildID, channelID string) entities.Channel {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	if stat, ok := entities.Guilds.DB[guildID].GetChannelStats()[channelID]; !ok {
		return stat
	}

	return entities.Channel{}
}

// SetGuildChannelStat sets a guild's channel stat in-memory
func SetGuildChannelStat(guildID string, channel entities.Channel) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].AssignToChannelStats(channel.GetChannelID(), channel)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("channelStats", entities.Guilds.DB[guildID].GetChannelStats())
}
