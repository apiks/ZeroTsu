package db

import (
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildFeedChecks returns the guild's feed checks from in-memory
func GetGuildFeedChecks(guildID string) []entities.FeedCheck {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	g := entities.Guilds.DB[guildID]
	entities.Guilds.RUnlock()

	return g.GetFeedChecks()
}

// AddGuildFeedChecks adds a target guild's feed checks in-memory
func AddGuildFeedChecks(guildID string, feedChecks []entities.FeedCheck, delete ...bool) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	if len(delete) == 0 {
		var exists bool
		for _, feedCheck := range feedChecks {
			for i, guildFeedCheck := range entities.Guilds.DB[guildID].GetFeedChecks() {
				if guildFeedCheck == feedCheck {
					entities.Guilds.DB[guildID].AssignToFeedChecks(i, feedCheck)
					exists = true
					break
				}
			}

			if !exists {
				entities.Guilds.DB[guildID].AppendToFeedChecks(feedCheck)
			}
		}
	} else {
		deleteGuildFeedChecks(guildID, feedChecks)
	}
	g := entities.Guilds.DB[guildID]
	entities.Guilds.Unlock()

	g.WriteData("rssThreadCheck", g.GetFeedChecks())
}

// AddGuildFeedCheck adds a target guild's feed check in-memory
func AddGuildFeedCheck(guildID string, feedCheck entities.FeedCheck, delete ...bool) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	if len(delete) == 0 {
		var exists bool
		for i, guildFeedCheck := range entities.Guilds.DB[guildID].GetFeedChecks() {
			if guildFeedCheck == feedCheck {
				entities.Guilds.DB[guildID].AssignToFeedChecks(i, feedCheck)
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
	g := entities.Guilds.DB[guildID]
	entities.Guilds.Unlock()

	g.WriteData("rssThreadCheck", g.GetFeedChecks())
}

// SetGuildFeedCheck sets a target guild's feed check in-memory
func SetGuildFeedCheck(guildID string, feedCheck entities.FeedCheck, delete ...bool) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	if len(delete) == 0 {
		var exists bool
		for i, guildFeedCheck := range entities.Guilds.DB[guildID].GetFeedChecks() {
			if guildFeedCheck == feedCheck {
				entities.Guilds.DB[guildID].AssignToFeedChecks(i, feedCheck)
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
	g := entities.Guilds.DB[guildID]
	entities.Guilds.Unlock()

	g.WriteData("rssThreadCheck", g.GetFeedChecks())
}

// deleteGuildFeedChecks safely deletes multiple feed checks from the feedChecks slice
func deleteGuildFeedChecks(guildID string, feedChecks []entities.FeedCheck) {
	for _, feedCheck := range feedChecks {
		for i, guildFeedCheck := range entities.Guilds.DB[guildID].GetFeedChecks() {
			if guildFeedCheck == feedCheck {
				entities.Guilds.DB[guildID].RemoveFromFeedChecks(i)
				break
			}
		}
	}
}

// deleteGuildFeedCheck safely deletes a feed check from the feedChecks slice
func deleteGuildFeedCheck(guildID string, feedCheck entities.FeedCheck) {
	for i, guildFeedCheck := range entities.Guilds.DB[guildID].GetFeedChecks() {
		if guildFeedCheck == feedCheck {
			entities.Guilds.DB[guildID].RemoveFromFeedChecks(i)
			break
		}
	}
}
