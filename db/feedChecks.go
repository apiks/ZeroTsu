package db

import (
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildFeedChecks returns the guild's feed checks from in-memory
func GetGuildFeedChecks(guildID string) []*entities.FeedCheck {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetFeedChecks()
}

// SetGuildFeedChecks sets a target guild's feed checks in the redis instance or in-memory
func SetGuildFeedChecks(guildID string, feedChecks []*entities.FeedCheck) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].SetFeedChecks(feedChecks)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("rssThreadCheck", entities.Guilds.DB[guildID].GetFeedChecks())
}

// SetGuildFeedCheck sets a target guild's feed check in the redis instance or in-memory
func SetGuildFeedCheck(guildID string, feedCheck entities.FeedCheck, delete ...bool) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	if len(delete) == 0 {
		var exists bool
		for _, guildFeedCheck := range entities.Guilds.DB[guildID].GetFeedChecks() {
			if guildFeedCheck == nil {
				continue
			}

			if *guildFeedCheck == feedCheck {
				*guildFeedCheck = feedCheck
				exists = true
				break
			}
		}

		if !exists {
			entities.Guilds.DB[guildID].AppendToFeedChecks(feedCheck)
		}
	} else {
		deleteGuildFeedCheck(guildID, feedCheck)
	}

	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("rssThreadCheck", entities.Guilds.DB[guildID].GetFeedChecks())

	return nil
}

// deleteGuildFeedCheck safely deletes a feed Check from the feedChecks slice
func deleteGuildFeedCheck(guildID string, feedCheck entities.FeedCheck) {
	for i, guildFeedCheck := range entities.Guilds.DB[guildID].GetFeedChecks() {
		if guildFeedCheck == nil {
			continue
		}

		if *guildFeedCheck == feedCheck {
			entities.Guilds.DB[guildID].RemoveFromFeedChecks(i)
			break
		}
	}
}
