package db

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/entities"
)

// GetGuildVoteInfo returns a guild's voteInfo map from in-memory
func GetGuildVoteInfo(guildID string) map[string]*entities.VoteInfo {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetVoteInfoMap()
}

// SetGuildVoteInfo sets a guild's VoteInfo in-memory
func SetGuildVoteInfo(guildID string, voteInfo map[string]*entities.VoteInfo) {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()
	entities.Guilds.DB[guildID].SetVoteInfoMap(voteInfo)
	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("voteInfo", entities.Guilds.DB[guildID].GetVoteInfoMap())
}

// SetGuildVoteInfoChannel sets a guild's vote info channel in-memory
func SetGuildVoteInfoChannel(guildID, messageID string, vote *entities.VoteInfo, deleteSlice ...bool) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	if len(deleteSlice) == 0 {
		if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetVoteInfoMap()) >= 200 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the vote limit (200) for this premium server.")
		} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetVoteInfoMap()) >= 50 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the vote limit (50) for this server. Please wait for some to be removed or increase them to 200 by upgrading to a premium server at <https://patreon.com/animeschedule>")
		}
	}

	if len(deleteSlice) == 0 {
		var exists bool
		for _, guildVote := range entities.Guilds.DB[guildID].GetVoteInfoMap() {
			if guildVote == nil {
				continue
			}

			if guildVote.GetUser() == vote.GetUser() &&
				guildVote.GetChannel() == vote.GetChannel() &&
				guildVote.GetChannelType() == vote.GetChannelType() {
				*guildVote = *vote
				exists = true
				break
			}
		}

		if !exists {
			entities.Guilds.DB[guildID].GetVoteInfoMap()[messageID] = vote
		}
	} else {
		delete(entities.Guilds.DB[guildID].GetVoteInfoMap(), messageID)
	}

	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("voteInfo", entities.Guilds.DB[guildID].GetVoteInfoMap())

	return nil
}
